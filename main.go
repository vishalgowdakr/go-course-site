package main

import (
	"course-site/templates"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Chapter struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

type Unit struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Chapters    []Chapter `json:"chapters"`
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
	m.path = cwd + "/lessons"
	entries, err := os.ReadDir(m.path)
	if err != nil {
		log.Fatal(err)
	}
	var units []Unit
	for _, entry := range entries {
		if entry.IsDir() {
			var chapters []Chapter
			m.max_units++
			subentries, err := os.ReadDir(m.path + entry.Name())
			if err != nil {
				log.Fatal(err)
			}
			var count int = 0
			for _, subentry := range subentries {
				if split := strings.Split(subentry.Name(), "."); split[len(split)-1] == "md" {
					count++
				} else if subentry.Name() == "info.json" {
					infoFIle, err := os.ReadFile(m.path + entry.Name() + subentry.Name())
					if err != nil {
						log.Fatal(err)
					}
					var chapter Chapter
					err = json.Unmarshal(infoFIle, &chapter)
					if err != nil {
						log.Fatal(err)
					}
					chapter.Path = m.path + entry.Name() + subentry.Name()
					chapters = append(chapters, chapter)
				}
			}
			m.chapters_count = append(m.chapters_count, count)
			infoFIle, err := os.ReadFile(m.path + entry.Name())
			if err != nil {
				log.Fatal(err)
			}
			var unit Unit
			err = json.Unmarshal(infoFIle, &unit)
			if err != nil {
				log.Fatal(err)
			}
			unit.Chapters = chapters
			units = append(units, unit)
		}

	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Static("public/", "public")
	e.GET("", func(c echo.Context) error {
		return templates.Home().Render(c.Request().Context(), c.Response().Writer)
	})
	/* e.GET("to/:unit/:chapter", func(c echo.Context) error {
		unit := c.Param("unit")
		chapter := c.Param("chapter")
	}) */
	e.StdLogger.Fatal(e.Start("localhost:6996"))
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
