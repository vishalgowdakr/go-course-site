package main

import (
	"context"
	"course-site/templates"
	"os"
	"path/filepath"

	// "github.com/a-h/templ"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/labstack/echo/v4"
)

type Model struct {
	Current_state int
	Paths         []string
	Index         int
	Is_first_page bool
	Is_last_page  bool
	Err           error
}

const (
	home_page = iota
	lesson_page
)

type Msg interface{}

type Cmd func(Model) Msg

type NextPageMsg struct{}
type PrevPageMsg struct{}
type GotoMsg struct {
	page  int
	index int
}
type ErrorMsg struct {
	e error
}

func mdToHtml(md []byte) []byte {
	extentions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extentions)
	doc := p.Parse(md)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return markdown.Render(doc, renderer)
}

func (m Model) Init() {
	m.Current_state = home_page
	if len(os.Args) > 1 {
		base_path := os.Args[1]
		files, err := os.ReadDir(base_path)
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			if !file.IsDir() {
				filePath := filepath.Join(base_path, file.Name())
				m.Paths = append(m.Paths, filePath)
			}
		}
	}
	m.Index = -1
	m.Is_first_page = false
	m.Is_last_page = false
}

func (m Model) Update(msg Msg) Model {
	switch msg := msg.(type) {
	case NextPageMsg:
		if m.Index == len(m.Paths)-1 {
			m.Is_last_page = true
		}
		if !m.Is_last_page {
			m.Index++
		}
	case PrevPageMsg:
		if m.Index == 1 {
			m.Is_first_page = true
		}
		if !m.Is_first_page {
			m.Index--
		}
	case GotoMsg:
		m.Current_state = msg.page
		m.Index = msg.index
	case ErrorMsg:
		m.Err = msg.e
	}
	return m
}

/* func (m model) View() templ.Component {
	if m.current_state == home_page {
		//return the home_page
	}
	if m.current_state == lesson_page {
		//return the lesson page based on the index
	}
} */

func main() {
	model := Model{}
	model.Init()
	c := context.WithValue(context.Background(), "title", "Hello, World!")
	component := templates.Base(c)
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return component.Render(context.Background(), c.Response().Writer)
	})
	e.Static("/public", "public")
	e.Logger.Fatal(e.Start(":6996"))
}
