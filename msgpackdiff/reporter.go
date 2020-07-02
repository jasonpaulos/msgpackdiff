package msgpackdiff

type layer struct {
	object       *MsgpObject
	currentIndex int
	currentKey   string
}

type difference struct {
	isDeletion bool
	object     MsgpObject
	path       []layer
}

type Reporter struct {
	Brief       bool
	path        []layer
	differences []difference
}

func (r *Reporter) EnterMap(mapObject MsgpObject) {
	mapLayer := layer{
		object: &mapObject,
	}
	r.path = append(r.path, mapLayer)
}

func (r *Reporter) SetKey(index int, key string) {
	r.path[len(r.path)-1].currentIndex = index
	r.path[len(r.path)-1].currentKey = key
}

func (r *Reporter) LeaveMap() {
	r.path = r.path[:len(r.path)-1]
}

func (r *Reporter) EnterArray(arrayObject MsgpObject) {
	arrayLayer := layer{
		object: &arrayObject,
	}
	r.path = append(r.path, arrayLayer)
}

func (r *Reporter) SetIndex(index int) {
	r.path[len(r.path)-1].currentIndex = index
}

func (r *Reporter) LeaveArray() {
	r.path = r.path[:len(r.path)-1]
}

func (r *Reporter) LogDeletion(deleted MsgpObject) {
	d := difference{
		isDeletion: true,
		object:     deleted,
		path:       append([]layer(nil), r.path...),
	}
	r.differences = append(r.differences, d)
}

func (r *Reporter) LogAddition(added MsgpObject) {
	d := difference{
		isDeletion: false,
		object:     added,
		path:       append([]layer(nil), r.path...),
	}
	r.differences = append(r.differences, d)
}

func (r *Reporter) LogDifference(old MsgpObject, new MsgpObject) {
	r.LogDeletion(old)
	r.LogAddition(new)
}
