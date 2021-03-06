package apps

import (
	"github.com/gopherjs/gopherjs/js"
	. "github.com/influx6/gu/design"
	. "github.com/influx6/gu/trees"
	. "github.com/influx6/gu/trees/elems"
	. "github.com/influx6/gu/trees/events"
	. "github.com/influx6/gu/trees/property"
)

var _ = Resource(func() {

	DoTitle("Hello App")

	DoMarkup(func() *Markup {
		return Div(
			CSS(`
				${
					width:100%;
					height: 100%;
				}

				$ h1 {
					font-size: {{ .Size }};
					text-align: center;
					margin: 20% 0  0 5%;
				}

				$ span {
					text-align: center;
					font-weight: bold;
					margin: 0 0 0 50%;
					color: rgba(0,0,0,0.5);
				}

				$ h1::after{
					content:"!";
					display: inline-block;
				}

				$ h1:hover::after{
					content:"*";
					color: red;
					display: inline-block;
				}

			`, struct{ Size string }{Size: "130px"}),
			IDAttr("hello"),
			Header1(
				Text("Hello"),
				ClickEvent(func(ev EventObject, tree *Markup) {
					js.Global.Call("alert", "I just got clicked, Yaay!!!")
				}, ""),
			),
			Span(Text("Click me")),
		)
	}, "", false, false)

})
