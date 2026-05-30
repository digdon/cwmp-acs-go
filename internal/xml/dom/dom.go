package dom

import (
	"bytes"
	stdxml "encoding/xml"
)

type Attr struct {
	Name  string
	Value string
}

type Element struct {
	Name     string
	Attrs    []Attr
	Children []*Element
	Text     string
}

func NewElement(name string) *Element {
	return &Element{Name: name}
}

func (e *Element) AddAttr(name, value string) *Element {
	e.Attrs = append(e.Attrs, Attr{Name: name, Value: value})
	return e
}

func (e *Element) SetText(text string) *Element {
	e.Text = text
	return e
}

// AddChild appends child to e and returns child for further building.
func (e *Element) AddChild(child *Element) *Element {
	e.Children = append(e.Children, child)
	return child
}

type Document struct {
	Root *Element
}

func NewDocument(root *Element) *Document {
	return &Document{Root: root}
}

func (d *Document) Serialize() (string, error) {
	var buf bytes.Buffer
	enc := stdxml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := encode(enc, d.Root); err != nil {
		return "", err
	}
	if err := enc.Flush(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func encode(enc *stdxml.Encoder, e *Element) error {
	start := stdxml.StartElement{Name: stdxml.Name{Local: e.Name}}
	for _, a := range e.Attrs {
		start.Attr = append(start.Attr, stdxml.Attr{Name: stdxml.Name{Local: a.Name}, Value: a.Value})
	}
	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	if e.Text != "" {
		if err := enc.EncodeToken(stdxml.CharData(e.Text)); err != nil {
			return err
		}
	}
	for _, child := range e.Children {
		if err := encode(enc, child); err != nil {
			return err
		}
	}
	return enc.EncodeToken(start.End())
}
