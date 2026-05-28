package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"cwmp-acs/internal/cwmp"
)

func ParseSOAPEnvelope(xmlBody []byte) (*ParsedEnvelope, error) {
	decoder := xml.NewDecoder(bytes.NewReader(xmlBody))

	parsed := &ParsedEnvelope{
		Namespaces: map[string]string{},
	}

	var inHeader bool
	var inBody bool
	var headerStack []*SOAPElement
	var bodyStack []*SOAPElement

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if parsed.Envelope.Name.Local == "" {
				if t.Name.Local != "Envelope" || t.Name.Space != SOAPENV_NS_URL {
					return nil, fmt.Errorf("expected SOAP Envelope, got %s (%s)", t.Name.Local, t.Name.Space)
				}
				parsed.Envelope = SOAPElement{Name: t.Name, Attrs: append([]xml.Attr(nil), t.Attr...)}
				collectNamespaces(t.Attr, parsed.Namespaces)
				continue
			}

			if t.Name.Space == SOAPENV_NS_URL && t.Name.Local == "Header" {
				if parsed.Header.Name.Local != "" {
					return nil, fmt.Errorf("multiple SOAP Header elements found")
				}
				parsed.Header = SOAPElement{Name: t.Name, Attrs: append([]xml.Attr(nil), t.Attr...)}
				inHeader = true
				headerStack = []*SOAPElement{&parsed.Header}
				continue
			}

			if inHeader {
				child := SOAPElement{Name: t.Name, Attrs: append([]xml.Attr(nil), t.Attr...)}
				parent := headerStack[len(headerStack)-1]
				parent.Children = append(parent.Children, child)
				headerStack = append(headerStack, &parent.Children[len(parent.Children)-1])
				continue
			}

			if t.Name.Space == SOAPENV_NS_URL && t.Name.Local == "Body" {
				if parsed.Body.Name.Local != "" {
					return nil, fmt.Errorf("multiple SOAP Body elements found")
				}
				parsed.Body = SOAPElement{Name: t.Name, Attrs: append([]xml.Attr(nil), t.Attr...)}
				inBody = true
				bodyStack = []*SOAPElement{&parsed.Body}
				continue
			}

			if inBody {
				child := SOAPElement{Name: t.Name, Attrs: append([]xml.Attr(nil), t.Attr...)}
				parent := bodyStack[len(bodyStack)-1]
				parent.Children = append(parent.Children, child)
				bodyStack = append(bodyStack, &parent.Children[len(parent.Children)-1])
				continue
			}

		case xml.CharData:
			if inHeader && len(headerStack) > 0 {
				text := strings.TrimSpace(string(t))
				if text != "" {
					current := headerStack[len(headerStack)-1]
					current.Text += text
				}
			}
			if inBody && len(bodyStack) > 0 {
				text := strings.TrimSpace(string(t))
				if text != "" {
					current := bodyStack[len(bodyStack)-1]
					current.Text += text
				}
			}

		case xml.EndElement:
			if inHeader {
				if len(headerStack) > 0 {
					headerStack = headerStack[:len(headerStack)-1]
				}

				if len(headerStack) == 0 {
					headerStack = nil
					inHeader = false
				}
			}
			if inBody {
				if len(bodyStack) > 0 {
					bodyStack = bodyStack[:len(bodyStack)-1]
				}

				if len(bodyStack) == 0 {
					bodyStack = nil
					inBody = false
				}
			}
		}
	}

	// Need to do some checks to see if we have at least the basic structure of a CWMP message
	// (namespaces, header, body [with only a single child], etc)

	if parsed.Envelope.Name.Local == "" {
		return nil, fmt.Errorf("SOAP envelope not found")
	}
	if parsed.Header.Name.Local == "" {
		return nil, fmt.Errorf("SOAP header not found")
	}
	if parsed.Body.Name.Local == "" {
		return nil, fmt.Errorf("SOAP body not found")
	}

	return parsed, nil
}

func collectNamespaces(attrs []xml.Attr, out map[string]string) {
	for _, attr := range attrs {
		if attr.Name.Space == "xmlns" {
			out[attr.Name.Local] = attr.Value
			continue
		}
		if attr.Name.Space == "" && attr.Name.Local == "xmlns" {
			out[""] = attr.Value
		}
	}
}

func ParseCPEHeader(headerElem SOAPElement, cwmpNS string) cwmp.CwmpHeader {
	header := cwmp.CwmpHeader{}

	for _, child := range headerElem.Children {
		// if child.Name.Space != cwmpNS {
		// 	header.Unknown = append(header.Unknown, child)
		// 	continue
		// }

		switch child.Name.Local {
		case "ID":
			header.ID = child.Text
		case "SessionTimeout":
			header.SessionTimeout = child.Text
		case "SupportedCWMPVersions":
			header.SupportedCWMPVersions = child.Text
		}
	}

	return header
}

func BuildNamespaceMap(namespaces map[string]string) map[NamespaceID]Namespace {
	result := make(map[NamespaceID]Namespace)

	for prefix, url := range namespaces {
		if id, ok := namespaceUrlToIDMap[url]; ok {
			result[id] = Namespace{Prefix: prefix, URL: url}
		}
	}

	return result
}
