package js

import (
	"fmt"
	"strings"

	"github.com/go-humble/detect"
	"github.com/gopherjs/gopherjs/js"
)

// CreateFragment returns a DocumentFragment with the given html dom
func CreateFragment(html string) *js.Object {

	//if we are not in a browser,dont do anything.
	if !detect.IsBrowser() {
		return nil
	}

	// div := doc.CreateElement("div")
	div := CreateElement("div")

	//build up the html right in the div
	SetInnerHTML(div, html)

	//create the document fragment
	fragment := CreateDocumentFragment()

	//add the nodes from the div into the fragment
	nodes := ChildNodeList(div)
	ContextAppendChild(fragment, nodes...)

	//unwrap all the special Text UnknownELement we are using
	// UnWrapSpecialTextElements(fragment)

	return fragment
}

// AddNodeIfNone uses the dom.Node.IsEqualNode method to check if not already exist and if so swap them else just add
// NOTE: bad way of doing it use it as last option
func AddNodeIfNone(dest, src *js.Object) {
	AddNodeIfNoneInList(dest, ChildNodeList(dest), src)
}

// AddNodeIfNoneInList checks a node in a node list if it finds an equal it replaces only else does nothing
func AddNodeIfNoneInList(dest *js.Object, against []*js.Object, with *js.Object) bool {
	for _, no := range against {
		if IsEqualNode(no, with) {
			// if no.IsEqualNode(with) {
			ReplaceNode(dest, with, no)
			return true
		}
	}
	//not matching, add it
	ContextAppendChild(dest, with)
	// dest.AppendChild(with)
	return false
}

// Patch takes a dom string and creates a documentfragment from it and patches a existing dom element that is supplied. This algorithim only ever goes one-level deep, its not performant
// WARNING: this method is specifically geared to dealing with the haiku.Tree dom generation
func Patch(fragment, live *js.Object, onlyReplace bool) {

	//if we are not in a browser,dont do anything.
	if !detect.IsBrowser() {
		return
	}

	if !live.Call("hasChildNodes").Bool() {
		// if the live element is actually empty, then just append the fragment which
		// actually appends the nodes within it efficiently
		ContextAppendChild(live, fragment)
		return
	}

	shadowNodes := ChildNodeList(fragment)
	liveNodes := ChildNodeList(live)

	// FIXED: instead of going through the children which may be many,

patchloop:
	for n, node := range shadowNodes {
		if node == nil || node == js.Undefined {
			continue
		}

		if node.Get("constructor") == js.Global.Get("Text") {
			if _, empty := EmptyTextNode(node); empty {
				ContextAppendChild(live, node)
				continue patchloop
			}

			var liveNodeAt *js.Object

			if n < len(liveNodes) {
				liveNodeAt = liveNodes[n]
			}

			if liveNodeAt == nil || liveNodeAt == js.Undefined {
				ContextAppendChild(live, node)
			} else {
				InsertBefore(live, liveNodeAt, node)
			}

			continue patchloop
		}

		//get the tagname
		tagname := GetTag(node)

		// get the basic attrs
		var id, hash, class, uid string

		// do we have 'id' attribute? if so its a awesome chance to simplify
		if HasAttribute(node, "id") {
			id = GetAttribute(node, "id")
		}

		if HasAttribute(node, "class") {
			id = GetAttribute(node, "class")
		}

		// lets check for the hash and uid, incase its a pure template based script
		if HasAttribute(node, "hash") {
			hash = GetAttribute(node, "hash")
		}

		if HasAttribute(node, "uid") {
			uid = GetAttribute(node, "uid")
		}

		// if we have no id,class, uid or hash, we digress to bad approach of using Node.IsEqualNode
		if allEmpty(id, hash, uid) {
			AddNodeIfNone(live, node)
			continue patchloop
		}

		// eliminate which ones are empty and try to use the non empty to get our target
		if allEmpty(hash, uid) {
			// is the id empty also then we know class is not or vise-versa
			if allEmpty(id) {
				// log.Printf("adding since class")
				// class is it and we only want those that match narrowing our set
				no := QuerySelectorAll(live, class)

				// if none found we add else we replace
				if len(no) <= 0 {
					ContextAppendChild(live, node)
				} else {
					// check the available sets and replace else just add it
					AddNodeIfNoneInList(live, no, node)
				}

			} else {
				// id is it and we only want one
				// log.Printf("adding since id")
				no := QuerySelector(live, fmt.Sprintf("#%s", id))

				// if none found we add else we replace
				if no == nil || no != js.Undefined {
					ContextAppendChild(live, node)
				} else {
					ReplaceNode(live, node, no)
				}
			}

			continue patchloop
		}

		// lets use our unique id to check for the element if it exists
		sel := fmt.Sprintf(`%s[uid='%s']`, strings.ToLower(tagname), uid)

		// we know hash and uid are not empty so we kick ass the easy way
		targets := QuerySelectorAll(live, sel)

		// if we are nil then its a new node add it and return
		if len(targets) == 0 {
			ContextAppendChild(live, node)
			continue patchloop
		}

		for _, target := range targets {
			if onlyReplace {
				ReplaceNode(live, node, target)
				continue
			}

			//if we are to be removed then remove the target
			if HasAttribute(node, "NodeRemoved") {
				tgName := node.Get("tagName").String()

				// If its a tag in our header kdis, attempt to remove from head.
				if headerKids[tgName] {
					dom := js.Global.Get("document").Call("querySelector", "head")
					RemoveChild(dom, node)
				}

				// If its a script attempt to remove from head and body.
				if tgName == "script" {
					body := js.Global.Get("document").Call("querySelector", "body")
					head := js.Global.Get("document").Call("querySelector", "head")

					RemoveChild(head, node)
					RemoveChild(body, node)
				}

				// Lastly attempt to remove from target itself.
				RemoveChild(target, target)
				continue
			}

			// if the target hash is exactly the same with ours skip it
			if GetAttribute(target, "hash") == hash {
				continue
			}

			nchildren := ChildNodeList(node)

			//if the new node has no children, then just replace it.
			if len(nchildren) <= 0 {
				ReplaceNode(live, node, target)
				continue
			}

			//here we are not be removed and we do have kids

			//cleanout all the targets text-nodes
			CleanAllTextNode(target)

			//so we got this dude, are we already one level deep ? if so swap else
			// run through the children with Patch
			// if level >= 1 {
			// live.ReplaceChild(node, target)
			attrs := Attributes(node)

			for key, value := range attrs {
				SetAttribute(target, key, value)
			}

			children := ChildNodeList(target)
			if len(children) == 0 {
				SetInnerHTML(target, "")
				ContextAppendChild(target, nchildren...)

				// continue patchloop
				continue
			}

			Patch(node, target, onlyReplace)
		}
	}
}

// allEmpty checks if all strings supplied are empty
func allEmpty(s ...string) bool {
	var state = true

	for _, so := range s {
		if so != "" {
			state = false
			return state
		}
	}

	return state
}
