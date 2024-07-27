package main

import (
	"bytes"
	"course-site/templates"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/a-h/templ"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/labstack/echo/v4"
)

type Unit struct {
	Title   string   `json:"name"`
	Lessons []Lesson `json:"lessons"`
	Length  int      `json:"length"`
	Index   int      `json:"index"`
}

type Lesson struct {
	Title string
	Body  []byte
	Path  string
}

type Model struct {
	mu           sync.RWMutex
	CurrentState int
	Index        int
	Units        []Unit
	IsFirstPage  bool
	IsLastPage   bool
	Err          error
	EchoContext  *echo.Context
	IsHxRequest  bool
}

type ModelHandler func(*Model, echo.Context) error

const (
	homePage = iota
	lessonPage
)

type Msg interface{}

type NextPageMsg struct{}
type PrevPageMsg struct{}
type GotoMsg struct {
	Page   int
	Unit   int
	Lesson int
}
type ErrorMsg struct {
	Err error
}

func withModel(handler ModelHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "must-revalidate")
		model, ok := c.Get("model").(*Model)
		if !ok {
			return fmt.Errorf("model not found in context")
		}
		err := handler(model, c)
		if err != nil {
			return fmt.Errorf("error rendering view: %w", err)
		}
		return nil
	}
}

func mdToHtml(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return markdown.Render(doc, renderer)
}

func (m *Model) Init(basePath string, c echo.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CurrentState = homePage
	m.Units = []Unit{}

	files, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("error reading base directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			unitPath := filepath.Join(basePath, file.Name())
			infoFile, err := os.ReadFile(filepath.Join(unitPath, "info.json"))
			if err != nil {
				log.Printf("Error reading info file for %s: %v", file.Name(), err)
				continue
			}

			// Remove potential trailing commas and whitespace
			infoFile = bytes.TrimSpace(infoFile)
			infoFile = bytes.TrimRight(infoFile, ",")
			if infoFile[len(infoFile)-1] != '}' {
				infoFile = append(infoFile, '}')
			}

			var unit Unit
			d := json.NewDecoder(bytes.NewReader(infoFile))
			d.DisallowUnknownFields()
			if err := d.Decode(&unit); err != nil {
				log.Printf("Error unmarshaling info file for %s: %v", file.Name(), err)
				log.Printf("Content of info.json: %s", string(infoFile))
				continue
			}

			// Read lesson files
			lessonFiles, err := os.ReadDir(unitPath)
			if err != nil {
				log.Printf("Error reading lesson directory for %s: %v", file.Name(), err)
				continue
			}

			for _, lessonFile := range lessonFiles {
				if filepath.Ext(lessonFile.Name()) == ".md" {
					lessonPath := filepath.Join(unitPath, lessonFile.Name())
					mdFile, err := os.ReadFile(lessonPath)
					if err != nil {
						log.Printf("Error reading lesson file %s: %v", lessonFile.Name(), err)
						continue
					}

					title := string(mdFile[:bytes.IndexByte(mdFile, '\n')])
					body := mdToHtml(mdFile)

					unit.Lessons = append(unit.Lessons, Lesson{
						Title: title,
						Body:  body,
						Path:  lessonPath,
					})
					unit.Length++
				}
			}

			m.Units = append(m.Units, unit)
			log.Printf("Added unit: %s with %d lessons", unit.Title, unit.Length)
		}
	}

	if len(m.Units) == 0 {
		return fmt.Errorf("no units found in the specified path")
	}

	log.Printf("Initialized with %d units", len(m.Units))
	m.IsFirstPage = true
	m.IsLastPage = false

	return nil
}

func (m *Model) Update(msg Msg) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.Units) == 0 {
		// Handle the case when there are no units
		m.Err = fmt.Errorf("no units available")
		return
	}

	switch msg := msg.(type) {
	case NextPageMsg:
		m.CurrentState = lessonPage
		if m.Index >= len(m.Units) {
			m.IsLastPage = true
			return
		}
		lastLen := len(m.Units[m.Index].Lessons)
		if m.Index == len(m.Units)-1 && m.Units[m.Index].Index == lastLen-1 {
			m.IsLastPage = true
			return
		}
		if m.Units[m.Index].Index == m.Units[m.Index].Length-1 {
			m.Index++
			if m.Index < len(m.Units) {
				m.Units[m.Index].Index = 0
			}
		} else {
			m.Units[m.Index].Index++
		}
		m.IsFirstPage = false

	case PrevPageMsg:
		m.CurrentState = lessonPage
		if m.Index == 0 && m.Units[m.Index].Index == 0 {
			m.IsFirstPage = true
			return
		}
		if m.Units[m.Index].Index == 0 {
			if m.Index > 0 {
				m.Index--
				m.Units[m.Index].Index = m.Units[m.Index].Length - 1
			}
		} else {
			m.Units[m.Index].Index--
		}
		m.IsLastPage = false

	case GotoMsg:
		if msg.Unit < len(m.Units) && msg.Lesson < len(m.Units[msg.Unit].Lessons) {
			m.Index = msg.Unit
			m.Units[msg.Unit].Index = msg.Lesson
			m.CurrentState = msg.Page
		} else {
			m.Err = fmt.Errorf("invalid unit or lesson index")
		}

	case ErrorMsg:
		m.Err = msg.Err
	}
}

func (m *Model) View() templ.Component {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch m.CurrentState {
	case homePage:
		return templates.Home()
	case lessonPage:
		if m.IsHxRequest {
			return templ.Raw(string(mdToHtml(m.Units[m.Index].Lessons[m.Units[m.Index].Index].Body)) + "<script>hljs.highlightAll();</script>")
		}
		return templates.Lessons()
	default:
		return templates.Home()
	}
}

func main() {
	e := echo.New()
	e.Use(setModel)

	e.GET("/", withModel(func(model *Model, c echo.Context) error {
		model.Update(GotoMsg{Page: homePage})
		return model.View().Render(c.Request().Context(), c.Response().Writer)
	}))

	e.GET("/lessons", withModel(func(model *Model, c echo.Context) error {
		model.Update(GotoMsg{Page: lessonPage, Unit: 0, Lesson: 0})
		return model.View().Render(c.Request().Context(), c.Response().Writer)
	}))

	e.GET("/lessons/next", withModel(func(model *Model, c echo.Context) error {
		if c.Request().Header.Get("HX-Request") == "true" {
			model.IsHxRequest = true
			defer func() {
				model.IsHxRequest = false
			}()
		}
		model.Update(NextPageMsg{})
		log.Print("Next page")
		log.Print(model.Index)
		log.Print(model.Units[model.Index].Index)
		return model.View().Render(c.Request().Context(), c.Response().Writer)
	}))

	e.GET("/lessons/prev", withModel(func(model *Model, c echo.Context) error {
		if c.Request().Header.Get("HX-Request") == "true" {
			model.IsHxRequest = true
			defer func() {
				model.IsHxRequest = false
			}()
		}
		model.Update(PrevPageMsg{})
		log.Print("Prev page")
		log.Print(model.Index)
		log.Print(model.Units[model.Index].Index)
		return model.View().Render(c.Request().Context(), c.Response().Writer)
	}))

	e.GET("/lessons/highlight", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})

	e.Static("/public", "public")

	e.Logger.Fatal(e.Start(":6996"))
}
