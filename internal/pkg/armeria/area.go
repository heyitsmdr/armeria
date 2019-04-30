package armeria

import (
	"sync"
)

type Area struct {
	Name  string  `json:"name"`
	Rooms []*Room `json:"rooms"`
	mux   sync.Mutex
}

type Room struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Coords      *Coords `json:"coords"`
	objects     []Object
	mux         sync.Mutex
}

type Coords struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
	I int `json:"-"`
}

type Location struct {
	AreaName string  `json:"area"`
	Coords   *Coords `json:"coords"`
}

func (a *Area) GetName() string {
	a.mux.Lock()
	defer a.mux.Unlock()
	return a.Name
}

func (a *Area) GetRoom(c *Coords) *Room {
	a.mux.Lock()
	defer a.mux.Unlock()

	for _, r := range a.Rooms {
		rc := r.GetCoords()
		if rc.X == c.X && rc.Y == c.Y && rc.Z == c.Z && rc.I == c.I {
			return r
		}
	}

	return nil
}

// GetMap returns
func (a *Area) GetMap() {

}

// OnCharacterEntered is called when the character is moved into the area (or logged in)
func (a *Area) OnCharacterEntered(c *Character, causedByLogin bool) {
	c.GetPlayer().clientActions.RenderMap(a.GetMinimap())
}

// OnCharacterLeft is called when the character left the area (or logged out)
func (a *Area) OnCharacterLeft(c *Character, causedByLogout bool) {

}

func (r *Room) GetCoords() *Coords {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.Coords
}

func (r *Room) SetTitle(title string) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.Title = title
}

func (r *Room) GetTitle() string {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.Title
}

func (r *Room) GetDescription() string {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.Description
}

func (r *Room) GetObjects() []Object {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.objects
}

func (r *Room) GetCharacters(except *Character) []*Character {
	r.mux.Lock()
	defer r.mux.Unlock()

	var returnChars []*Character

	for _, o := range r.objects {
		if o.GetType() == ObjectTypeCharacter {
			if except == nil || o.GetName() != except.GetName() {
				returnChars = append(returnChars, o.(*Character))
			}
		}
	}

	return returnChars
}

func (r *Room) AddObjectToRoom(obj Object) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.objects = append(r.objects, obj)
}

func (r *Room) RemoveObjectFromRoom(obj Object) bool {
	r.mux.Lock()
	defer r.mux.Unlock()

	for i, o := range r.objects {
		if o.GetType() == obj.GetType() && o.GetName() == obj.GetName() {
			r.objects[i] = r.objects[len(r.objects)-1]
			r.objects = r.objects[:len(r.objects)-1]
			return true
		}
	}

	return false
}

// OnCharacterEntered is called when the character is moved to the room (or logged in)
func (r *Room) OnCharacterEntered(c *Character, causedByLogin bool) {
	c.GetPlayer().clientActions.SetLocation(c.GetLocationData())
}

// OnCharacterLeft is called when the character left the room (or logged out)
func (r *Room) OnCharacterLeft(c *Character, causedByLogout bool) {

}
