package msgpackdiff

type DifferenceType int

const (
	Deletion DifferenceType = iota
	Addition
	Replacement
)

type Layer struct {
	Object       *MsgpObject
	CurrentIndex int
	CurrentKey   string
}

type Difference struct {
	Type   DifferenceType
	Object MsgpObject
	Path   []Layer
}

type Reporter struct {
	Brief       bool
	Path        []Layer
	Differences []Difference
}

func (r *Reporter) EnterMap(mapObject MsgpObject) {
	mapLayer := Layer{
		Object: &mapObject,
	}
	r.Path = append(r.Path, mapLayer)
}

func (r *Reporter) SetKey(index int, key string) {
	r.Path[len(r.Path)-1].CurrentIndex = index
	r.Path[len(r.Path)-1].CurrentKey = key
}

func (r *Reporter) LeaveMap() {
	r.Path = r.Path[:len(r.Path)-1]
}

func (r *Reporter) EnterArray(arrayObject MsgpObject) {
	arrayLayer := Layer{
		Object: &arrayObject,
	}
	r.Path = append(r.Path, arrayLayer)
}

func (r *Reporter) SetIndex(index int) {
	r.Path[len(r.Path)-1].CurrentIndex = index
}

func (r *Reporter) LeaveArray() {
	r.Path = r.Path[:len(r.Path)-1]
}

func (r *Reporter) LogDeletion(deleted MsgpObject) {
	d := Difference{
		Type:   Deletion,
		Object: deleted,
		Path:   append([]Layer(nil), r.Path...),
	}
	r.Differences = append(r.Differences, d)
}

func (r *Reporter) LogAddition(added MsgpObject) {
	d := Difference{
		Type:   Addition,
		Object: added,
		Path:   append([]Layer(nil), r.Path...),
	}
	r.Differences = append(r.Differences, d)
}

func (r *Reporter) LogChange(old MsgpObject, new MsgpObject) {
	path := append([]Layer(nil), r.Path...)
	deletion := Difference{
		Type:   Deletion,
		Object: old,
		Path:   path,
	}
	replacement := Difference{
		Type:   Replacement,
		Object: new,
		Path:   path,
	}
	r.Differences = append(r.Differences, deletion, replacement)
}
