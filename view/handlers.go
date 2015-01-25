package view

import (
	"fmt"
	"github.com/gophergala/golab/model"
	"html/template"
	"image"
	"image/jpeg"
	"net/http"
	"strconv"
	"time"
)

var params = struct {
	Title         string
	Width, Height int
	RunId         int64
}{AppTitle, ViewWidth, ViewHeight, time.Now().Unix()}

var playTempl = template.Must(template.New("t").Parse(play_html))

// The client's (browser's) view position inside the labyrinth image. This is the top-left point of the view.
var Pos image.Point

// init registers the http handlers.
func init() {
	http.HandleFunc("/", playHtmlHandle)
	http.HandleFunc("/runid", runIdHandle)
	http.HandleFunc("/img", imgHandle)
	http.HandleFunc("/clicked", clickedHandle)
	http.HandleFunc("/cheat", cheatHandle)
	http.HandleFunc("/new", newGameHandle)
}

// InitNew initializes a new game.
func InitNew() {
	Pos = image.Point{}
}

// playHtmlHandle serves the html page where the user can play.
func playHtmlHandle(w http.ResponseWriter, r *http.Request) {
	playTempl.Execute(w, params)
}

// runidHandle serves the running app id which changes if app is restarted
// (so browser clients can detect if app was restarted).
func runIdHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%d", params.RunId)
}

// imgHandle serves images of the player's view.
func imgHandle(w http.ResponseWriter, r *http.Request) {
	quality, err := strconv.Atoi(r.FormValue("quality"))
	if err != nil || quality < 0 || quality > 100 {
		quality = 70
	}

	// Center Gopher in view if possible
	gpos := model.Gopher.Pos
	rect := image.Rect(0, 0, ViewWidth, ViewHeight).Add(image.Pt(int(gpos.X)-ViewWidth/2, int(gpos.Y)-ViewHeight/2))

	// But needs correction at the edges of the view (it can't be centered)
	corr := image.Point{}
	if rect.Min.X < 0 {
		corr.X = -rect.Min.X
	}
	if rect.Min.Y < 0 {
		corr.Y = -rect.Min.Y
	}
	if rect.Max.X > model.LabWidth {
		corr.X = model.LabWidth - rect.Max.X
	}
	if rect.Max.Y > model.LabHeight {
		corr.Y = model.LabHeight - rect.Max.Y
	}
	rect = rect.Add(corr)

	model.Mutex.Lock()
	jpeg.Encode(w, model.LabImg.SubImage(rect), &jpeg.Options{quality})
	model.Mutex.Unlock()

	// Store the new view's position:
	Pos = rect.Min
}

// clickedHandle receives mouse click (mouse button pressed) events with mouse coordinates.
func clickedHandle(w http.ResponseWriter, r *http.Request) {
	x, err := strconv.Atoi(r.FormValue("x"))
	if err != nil {
		return
	}
	y, err := strconv.Atoi(r.FormValue("y"))
	if err != nil {
		return
	}

	// x, y are in the coordinate system of the client's view.
	// Translate them to the Labyrinth's coordinate system:
	model.ClickCh <- model.Click{Pos.X + x, Pos.Y + y}
}

// cheatHandle serves the whole image of the labyrinth
func cheatHandle(w http.ResponseWriter, r *http.Request) {
	model.Mutex.Lock()
	jpeg.Encode(w, model.LabImg, &jpeg.Options{70})
	model.Mutex.Unlock()
}

// newGameHandle signals to start a newgame
func newGameHandle(w http.ResponseWriter, r *http.Request) {
	// Use non-blocking send
	select {
	case model.NewGameCh <- 1:
	default:
	}
}
