package layout
type LayoutName string
const (
	LayoutSingle  LayoutName = "Single"
	LayoutExplore LayoutName = "Explore"
	LayoutQuery   LayoutName = "Query"
	LayoutFull    LayoutName = "Full"
)
var AllLayouts = []LayoutName{LayoutSingle, LayoutExplore, LayoutQuery, LayoutFull}
func NextLayout(current LayoutName) LayoutName {
	for i, name := range AllLayouts {
		if name == current {
			return AllLayouts[(i+1)%len(AllLayouts)]
		}
	}
	return LayoutSingle
}
func AutoLayout(width int) LayoutName {
	switch {
	case width >= 160:
		return LayoutFull
	case width >= 100:
		return LayoutExplore
	default:
		return LayoutSingle
	}
}
type Panes struct {
	Browser Renderable
	Editor  Renderable
	Results Renderable
}
func BuildLayout(name LayoutName, p Panes) *Container {
	switch name {
	case LayoutExplore:
		return NewSplit(Horizontal, 0.4,
			NewLeaf("Browser", p.Browser),
			NewLeaf("Results", p.Results),
		)
	case LayoutQuery:
		return NewSplit(Vertical, 0.5,
			NewLeaf("Editor", p.Editor),
			NewLeaf("Results", p.Results),
		)
	case LayoutFull:
		return NewSplit(Horizontal, 0.3,
			NewLeaf("Browser", p.Browser),
			NewSplit(Vertical, 0.5,
				NewLeaf("Editor", p.Editor),
				NewLeaf("Results", p.Results),
			),
		)
	default:
	}
	return nil
}