// Package gujs defines helper methods that use the underline gopherjs js package
// to inteface with the browser dom api.
package js

import (
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

// DOMObjectToList takes a jsobjects and returns a list of internal objects by calling the item method
func DOMObjectToList(o *js.Object) []*js.Object {
	var out []*js.Object
	length := o.Get("length").Int()
	for i := 0; i < length; i++ {
		out = append(out, o.Call("item", i))
	}
	return out
}

// ChildNodeList returns the nodes list of the childNodes of the js object
func ChildNodeList(o *js.Object) []*js.Object {
	return DOMObjectToList(o.Get("childNodes"))
}

// Attributes takes a js object and extracts the attribute lists from it
func Attributes(co *js.Object) map[string]string {
	o := co.Get("attributes")

	if o == nil || o == js.Undefined {
		return nil
	}

	attrs := map[string]string{}

	length := o.Get("length").Int()

	for i := 0; i < length; i++ {
		item := o.Call("item", i)
		attrs[item.Get("name").String()] = item.Get("value").String()
	}

	return attrs
}

// GetWindow returns the global object which is the window in the browser
func GetWindow() *js.Object {
	return js.Global
}

// GetDocument returns the document js.object from the global window object
func GetDocument() *js.Object {
	win := GetWindow()
	if win == nil || win == js.Undefined {
		return nil
	}
	return win.Get("document")
}

// CreateElement creates a dom element using the document html js.object
func CreateElement(tag string) *js.Object {
	doc := GetDocument()
	if doc == nil || doc == js.Undefined {
		return nil
	}
	return doc.Call("createElement", tag)
}

// CreateDocumentFragment creates a dom documentFragment using the document html js.object
func CreateDocumentFragment() *js.Object {
	doc := GetDocument()
	if doc == nil {
		return nil
	}
	return doc.Call("createDocumentFragment")
}

// var onlySpace = regexp

// EmptyTextNode returns two bool values, the first indicating if its a text node and the second indicating if the text node is empty
func EmptyTextNode(o *js.Object) (bool, bool) {
	if o.Get("nodeType").Int() == 3 {
		textContent := strings.TrimSpace(o.Get("textContent").String())
		if textContent != "" {
			return true, false
		}
		return true, true
	}
	return false, false
}

// CleanAllTextNode removes all texts nodes within the container root
func CleanAllTextNode(o *js.Object) {
	for _, to := range ChildNodeList(o) {
		if istx, isem := EmptyTextNode(to); istx {
			if !isem {
				o.Call("removeChild", to)
			}
		}
	}
}

// UnWrapSpecialTextElements takes a dom object and unwraps all the Text UnknownELement within the lists
func UnWrapSpecialTextElements(o *js.Object) {
	texts := QuerySelectorAll(o, "text")
	for _, to := range texts {
		parent := to.Get("parentNode")
		SpecialAppendChild(parent, to)
		parent.Call("removeChild", to)
	}
}

// SpecialAppendChild takes a list of objects and calls appendNode on the given object, but checks if the objects contain an unknownelement with a text tag then strip the tagname and only apply its content
func SpecialAppendChild(o *js.Object, osets ...*js.Object) {
	for _, onode := range osets {
		if strings.ToLower(onode.Get("tagName").String()) == "text" {
			SpecialAppendChild(o, ChildNodeList(onode)...)
			continue
		}
		o.Call("appendChild", onode)
	}
}

// InsertBefore inserts the inserto before the guage object with the target
func InsertBefore(target, guage, inserto *js.Object) {
	target.Call("insertBefore", inserto, guage)
}

// AppendChild takes a list of objects and calls appendNode on the given object
func AppendChild(o *js.Object, osets ...*js.Object) {
	for _, onode := range osets {
		o.Call("appendChild", onode)
	}
}

var headerKids = map[string]bool{
	"meta":  true,
	"link":  true,
	"title": true,
	"base":  true,
}

var bodyKids = map[string]bool{
	"scripts": true,
}

// ContextAppendChild takes a list of objects and calls appendNode on the given object
func ContextAppendChild(o *js.Object, osets ...*js.Object) {
	for _, onode := range osets {

		if doHeadAppend(onode) {
			continue
		}

		if scripts := QuerySelectorAll(onode, "scripts"); len(scripts) != 0 {
			for _, script := range scripts {
				doHeadAppend(script)
			}
		}

		o.Call("appendChild", onode)
	}
}

func doHeadAppend(onode *js.Object) bool {
	tagNameObject := onode.Get("tagName")
	if tagNameObject == nil || tagNameObject == js.Undefined {
		return false
	}

	tagName := strings.ToLower(tagNameObject.String())
	if !headerKids[tagName] {
		return false
	}

	uid := GetAttribute(onode, "uid")
	header := QuerySelector(GetDocument(), "head")
	body := QuerySelector(GetDocument(), "body")

	if headerKids[tagName] && header != nil && header != js.Undefined {
		possibleNode := QuerySelector(header, tagName+"[uid='"+uid+"']")
		if possibleNode != nil && possibleNode != js.Undefined {
			ReplaceNode(header, onode, possibleNode)
			return true
		}

		header.Call("appendChild", onode)
		return true
	}

	if bodyKids[tagName] && body != nil && body != js.Undefined {
		possibleNode := QuerySelector(body, tagName+"[uid='"+uid+"']")
		if possibleNode != nil && possibleNode != js.Undefined {
			ReplaceNode(body, onode, possibleNode)
			return true
		}

		body.Call("appendChild", onode)
		return true
	}

	return false
}

// RemoveChild takes a target and a list of children to remove
func RemoveChild(o *js.Object, co ...*js.Object) {
	for _, onode := range co {
		o.Get("parentNode").Call("removeChild", onode)
	}
}

// IsEqualNode returns a false/true if the nodes are equal in the eyes of the dom
func IsEqualNode(newNode, oldNode *js.Object) bool {
	return oldNode.Call("isEqualNode", newNode).Bool()
}

// ReplaceNode replaces two unequal nodes using their parents
func ReplaceNode(target, newNode, oldNode *js.Object) {
	if newNode == oldNode {
		return
	}
	target.Call("replaceChild", newNode, oldNode)
}

// QuerySelectorAll returns the result of querySelectorAll on an object
func QuerySelectorAll(o *js.Object, sel string) []*js.Object {
	if sad := o.Get("querySelectorAll"); sad == nil || sad == js.Undefined {
		return nil
	}

	return DOMObjectToList(o.Call("querySelectorAll", sel))
}

// QuerySelector returns the result of querySelector on an object
func QuerySelector(o *js.Object, sel string) *js.Object {
	return o.Call("querySelector", sel)
}

// GetTag returns the tag of the js object
func GetTag(o *js.Object) string {
	return o.Get("tagName").String()
}

// GetAttribute returns a string if a key exists using the jsobject
func GetAttribute(o *js.Object, key string) string {
	return o.Call("getAttribute", key).String()
}

// HasAttribute returns true/false if a key exists using the jsobject
func HasAttribute(o *js.Object, key string) bool {
	return o.Call("hasAttribute", key).Bool()
}

// SetAttribute calls setAttribute on the js object with the value and key
func SetAttribute(o *js.Object, key string, value string) {
	o.Call("setAttribute", key, value)
}

// SetInnerHTML calls the innerHTML setter with the given string
func SetInnerHTML(o *js.Object, html string) {
	o.Set("innerHTML", html)
}
