package main

import (
	"course-site/templates"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Chapter struct {
	filename string
}

type Unit struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Chapters    []Chapter // This will be populated programmatically
}

type Model struct {
	path           string
	max_units      int
	max_chapters   int
	chapters_count []int //contains length of each unit as int array for easy acccessibility
	units          []Unit
}

func (m *Model) init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	m.path = cwd + "/lessons/"
	entries, err := os.ReadDir(m.path)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			var chapters []Chapter
			m.max_units++
			subentries, err := os.ReadDir(m.path + "/" + entry.Name())
			if err != nil {
				log.Fatal(err)
			}
			var count int = 0
			for _, subentry := range subentries {
				if split := strings.Split(subentry.Name(), "."); split[len(split)-1] == "md" {
					chapters = append(chapters, Chapter{
						filename: subentry.Name(),
					})
					count++
				}
			}

			infoFile, err := os.ReadFile(m.path + "/" + entry.Name() + "/info.json")
			if err != nil {
				log.Fatal(err)
			}
			var unit Unit
			err = json.Unmarshal(infoFile, &unit)
			if err != nil {
				log.Fatal(err)
			}

			// Assign the chapters we found to the unit
			unit.Chapters = chapters

			m.chapters_count = append(m.chapters_count, len(chapters))
			m.units = append(m.units, unit)
		}
	}
	m.max_chapters = len(m.units[m.max_units-1].Chapters)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Static("public/", "public")

	m := Model{}
	m.init()

	e.GET("", func(c echo.Context) error {
		return templates.Home().Render(c.Request().Context(), c.Response().Writer)
	})

	e.GET("to/:unit/:chapter", func(c echo.Context) error {
		fullpage := false
		if !isHtmxReq(c) {
			fullpage = true
		}
		unit := c.Param("unit")
		unitInt, err := strToInt(unit)
		if err != nil {
			return redirectToHome(c)
		}
		chapter := c.Param("chapter")
		chapterInt, err := strToInt(chapter)
		if err != nil {
			return redirectToHome(c)
		}
		nextUrl, err := nextChapter(&m, unitInt, chapterInt)
		if err != nil {
			return redirectToHome(c)
		}
		prevUrl, err := prevChapter(&m, unitInt, chapterInt)
		if err != nil {
			return redirectToHome(c)
		}
		content, status := goTo(&m, unitInt, chapterInt)
		c.Response().Status = status
		return templates.Lessons(fullpage, content+"<script>hljs.highlightAll();</script>", prevUrl, nextUrl).Render(c.Request().Context(), c.Response().Writer)
	})
	e.StdLogger.Fatal(e.Start("0.0.0.0:10000"))
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

func strToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func goTo(m *Model, unit int, chapter int) (string, int) {
	// Check if the unit is valid
	if unit < 1 || unit > m.max_units {
		return "Invalid unit number", http.StatusBadRequest
	}
	// Adjust for zero-based indexing
	unitIndex := unit - 1
	// Check if the chapter is valid for this unit
	if chapter < 1 || chapter > m.chapters_count[unitIndex] {
		return "Invalid chapter number", http.StatusBadRequest
	}
	// Adjust for zero-based indexing
	chapterIndex := chapter - 1
	// Get the filename for the requested chapter
	filename := m.units[unitIndex].Chapters[chapterIndex].filename
	// Construct the full path to the file
	filepath := m.path + "/" + m.units[unitIndex].Name + "/" + filename
	// Read the file content
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "Error reading chapter content", http.StatusInternalServerError
	}
	// Convert markdown to HTML
	htmlContent := string(mdToHtml(content))
	return htmlContent, http.StatusOK
}

func nextChapter(m *Model, unit int, chapter int) (string, error) {
	// Check if current unit and chapter are valid
	if unit < 1 || unit > m.max_units ||
		chapter < 1 || chapter > m.chapters_count[unit-1] {
		return "", fmt.Errorf("invalid unit or chapter")
	}

	// Check if there's a next chapter in the current unit
	if chapter < m.chapters_count[unit-1] {
		return fmt.Sprintf("/to/%d/%d", unit, chapter+1), nil
	}

	// If we're at the last chapter of the current unit, move to the next unit
	if unit < m.max_units {
		return fmt.Sprintf("/to/%d/%d", unit+1, 1), nil
	}

	// If we're at the last chapter of the last unit, return the current URL
	return fmt.Sprintf("/to/%d/%d", unit, chapter), nil
}

func prevChapter(m *Model, unit int, chapter int) (string, error) {
	// Check if current unit and chapter are valid
	if unit < 1 || unit > m.max_units ||
		chapter < 1 || chapter > m.chapters_count[unit-1] {
		return "", fmt.Errorf("invalid unit or chapter")
	}

	// Check if there's a previous chapter in the current unit
	if chapter > 1 {
		return fmt.Sprintf("/to/%d/%d", unit, chapter-1), nil
	}

	// If we're at the first chapter of the current unit, move to the previous unit
	if unit > 1 {
		prevUnit := unit - 1
		return fmt.Sprintf("/to/%d/%d", prevUnit, m.chapters_count[prevUnit-1]), nil
	}

	// If we're at the first chapter of the first unit, return the current URL
	return fmt.Sprintf("/to/%d/%d", unit, chapter), nil
}

func isHtmxReq(c echo.Context) bool {
	if c.Request().Header.Get("Hx-Request") == "true" {
		return true
	}
	return false
}

func redirectToHome(c echo.Context) error {
	c.Response().Status = http.StatusBadRequest
	return templates.Home().Render(c.Request().Context(), c.Response().Writer)
}
