package webstyle

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"text/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.seankhliao.com/webstyle/picture"
)

var (
	//go:embed layout.tpl
	layoutTpl string
	//go:embed base.css
	baseCss string
	//go:embed compact.css
	compactCss string

	templateBase    = template.Must(template.New("basecss").Parse(baseCss))
	templateCompact = template.Must(template.New("basecss").Parse(baseCss + compactCss))

	TemplateFull    = template.Must(templateBase.New("").Parse(layoutTpl))
	TemplateCompact = template.Must(templateCompact.New("").Parse(layoutTpl))

	defaultMarkdown = goldmark.New(
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.Table,
			extension.TaskList,
			picture.Picture,
		),
		goldmark.WithParserOptions(
			parser.WithHeadingAttribute(), // {#some-id}
			parser.WithAutoHeadingID(),    // based on heading
		),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
)

type Data struct {
	Main string

	// Optional
	Style    string
	Title    string // defaults to h1
	Subtitle string // defaults to h2
	Desc     string // defaults to subtitle
	Head     string
	GTM      string
	URL      string
}

type Renderer struct {
	Markdown goldmark.Markdown
	Template *template.Template
}

func NewRenderer(t *template.Template) Renderer {
	return Renderer{
		Markdown: defaultMarkdown,
		Template: t,
	}
}

func (r Renderer) Render(w io.Writer, src io.Reader, d Data) error {
	b, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	node := r.Markdown.Parser().Parse(text.NewReader(b))
	for n := node.FirstChild(); n != nil; n = n.NextSibling() {
		if hd, ok := n.(*ast.Heading); ok {
			if hd.Level == 1 && d.Title == "" {
				d.Title = string(hd.Text(b))
			} else if hd.Level == 2 {
				d.Subtitle = string(hd.Text(b))
				if d.Desc == "" {
					d.Desc = d.Subtitle
				}
			}
		}
	}
	var mdBuf bytes.Buffer
	err = r.Markdown.Renderer().Render(&mdBuf, b, node)
	if err != nil {
		return fmt.Errorf("render markdown: %w", err)
	}
	d.Main = mdBuf.String() + d.Main
	err = r.Template.Execute(w, d)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	return nil
}

func (r Renderer) RenderBytes(src []byte, d Data) ([]byte, error) {
	var buf bytes.Buffer
	err := r.Render(&buf, bytes.NewReader(src), d)
	return buf.Bytes(), err
}
