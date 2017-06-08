package main

import (
 	"github.com/ojrac/opensimplex-go"
	"strconv"
	"strings"
	"fmt"
	"github.com/fogleman/gg"
	"math"
	"bytes"
	"image/png"
	"github.com/gorilla/mux"
	"net/http"
	"log"
)

type Noise struct {
	Noise *opensimplex.Noise
}

func (n *Noise) getNoise(x float64, y float64) float64{
	//fmt.Println()
	//fmt.Printf( "%v\t%v\n", x, y)
	//fmt.Printf("%0.4f\n", n.Noise.Eval2(x, y))
	//fmt.Printf("%0.4f\n", n.Noise.Eval2(x, y) / 2 + 0.5)

	return n.Noise.Eval2(x, y) / 2 + 0.5
}

//var noise1, noise2 *Noise

func newNoise(seedString string) *Noise{
	var seed, err = strconv.ParseInt(strings.ToLower(seedString), 36, 64)

	if err != nil {
		fmt.Printf("ERROR: Seed %s is an invalid seed.\n", seedString)
		//os.Exit(1)
	}

	return &Noise{opensimplex.NewWithSeed(seed)}
}



func rescale(fromBegin, fromEnd, toBegin, toEnd, x float64) float64{
	t := (x-fromBegin) / (fromEnd-fromBegin)
	v := toBegin + (toEnd - toBegin) * t
	return v
}

func getElevation(x, y int, tr *TileRequest) float64 {
	nx := (float64(x*4) / math.Pow(2.0, float64(tr.z+1)))/float64(width)
	ny := (float64(y*4) / math.Pow(2.0, float64(tr.z+1)))/float64(height)

	//fmt.Printf("%v\t%v\n", x, y)
	//fmt.Printf("%0.4f\t%0.4f\n", float64(x/width), float64(y/height))
	//fmt.Printf("%0.4f\t%0.4f\n", nx, ny)

	e := 1.00 * tr.noise1.getNoise(1 * nx, 1 * ny) +
		0.50 * tr.noise1.getNoise(2 * nx, 2 * ny) +
		0.25 * tr.noise1.getNoise(4 * nx, 4 * ny) +
		0.13 * tr.noise1.getNoise(8 * nx, 8 * ny) +
		0.06 * tr.noise1.getNoise(16 * nx, 16 * ny) +
		0.03 * tr.noise1.getNoise(32 * nx, 32 * ny)
	e /= 1.00+0.50+0.25+0.13+0.06+0.03
	e = math.Pow(e, 1.5)
	return rescale(0.25, 0.75, 0.0, 1.0, e)
}

func getMoisture(x, y int, tr *TileRequest) float64 {
	noise2 := tr.noise2
	nx := (float64(x*4) / math.Pow(2.0, float64(tr.z+1)))/float64(width)
	ny := (float64(y*4) / math.Pow(2.0, float64(tr.z+1)))/float64(height)

	m := 1.00 * noise2.getNoise(1 * nx, 1 * ny) +
		0.50 * noise2.getNoise(2 * nx, 2 * ny) +
		0.25 * noise2.getNoise(4 * nx, 4 * ny) +
		0.25 * noise2.getNoise(8 * nx, 8 * ny) +
		0.12 * noise2.getNoise(16 * nx, 16 * ny) +
		0.06 * noise2.getNoise(32 * nx, 32 * ny)
	m /= 1.00+0.50+0.25+0.25+0.12+0.06
	return rescale(0.25, 0.75, 0.0, 1.0, m)
}

func setTerrain(x, y int, e, e_prev, m float64, tr *TileRequest){

	//fmt.Printf("%v\t%v\t%0.4f\t%0.4f\n", x % width, y % width, e, m)
	if e < 0.1 {
		//OCEAN
		tr.G.SetRGB255(68, 68, 122)
		tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
		return
	}
	if e < 0.12 {
		//BEACH
		tr.G.SetRGB255(160, 144, 119)
		if (e_prev > e){
			tr.G.SetRGB255(120, 104, 79)
		}
		tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
		return
	}

	if e > 0.8 {
		if m < 0.1 {
			//SCORCHED
			tr.G.SetRGB255(85, 85, 85)
			if (e_prev > e){
				tr.G.SetRGB255(45, 45, 45)
			}
			tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
			return
		}
		if m < 0.2 {
			//BARE
			tr.G.SetRGB255(136, 136, 136)
			if (e_prev > e){
				tr.G.SetRGB255(96, 96, 96)
			}
			tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
			return
		}
		if m < 0.5 {
			//TUNDRA
			tr.G.SetRGB255(187, 187, 170)
			if (e_prev > e){
				tr.G.SetRGB255(147, 147, 130)
			}
			tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
			return
		}
		//SNOW
		tr.G.SetRGB255(221, 221, 228)
		if (e_prev > e){
			tr.G.SetRGB255(181, 181, 188)
		}
		tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
		return
	}
	if e > 0.6 {
		if m < 0.33 {
			//TEMPERATE_DESERT
			tr.G.SetRGB255(201, 210, 155)
			if (e_prev > e){
				tr.G.SetRGB255(161, 161, 115)
			}
			tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
			return
		}
		if m < 0.66 {
			//SHRUBLAND
			tr.G.SetRGB255(136, 153, 119)
			if (e_prev > e){
				tr.G.SetRGB255(96, 113, 79)
			}
			tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
			return
		}
		//TAIGA
		tr.G.SetRGB255(153, 170, 119)
		if (e_prev > e){
			tr.G.SetRGB255(113, 130, 79)
		}
		tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
		return
	}
	if e > 0.3 {
		if m < 0.16 {
			//TEMPERATE_DESERT
			tr.G.SetRGB255(201, 210, 155)
			if (e_prev > e){
				tr.G.SetRGB255(161, 170, 115)
			}
			tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
			return
		}
		if m < 0.50 {
			//GRASSLAND
			tr.G.SetRGB255(136, 170, 85)
			if (e_prev > e){
				tr.G.SetRGB255(96, 130, 45)
			}
			tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
			return
		}
		if m < 0.83 {
			//TEMPERATE_DECIDUOUS_FOREST
			tr.G.SetRGB255(103, 148, 89)
			if (e_prev > e){
				tr.G.SetRGB255(63, 108, 49)
			}
			tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
			return
		}
		//TEMPERATE_RAIN_FOREST
		tr.G.SetRGB255(68, 136, 85)
		if (e_prev > e){
			tr.G.SetRGB255(28, 96, 45)
		}
		tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
		return
	}

	if m < 0.16 {
		//SUBTROPICAL_DESERT
		tr.G.SetRGB255(210, 185, 139)
		if (e_prev > e){
			tr.G.SetRGB255(170, 145, 99)
		}
		tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
		return
	}
	if m < 0.50 {
		//GRASSLAND
		tr.G.SetRGB255(136, 170, 85)
		if (e_prev > e){
			tr.G.SetRGB255(96, 130, 45)
		}
		tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
		return
	}
	if m < 0.83 {
		//TROPICAL_SEASONAL_FOREST
		tr.G.SetRGB255(85, 153, 68)
		if (e_prev > e){
			tr.G.SetRGB255(45, 113, 28)
		}
		tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
		return
	}
	//TROPICAL_RAIN_FOREST
	tr.G.SetRGB255(51, 119, 85)
	if (e_prev > e){
		tr.G.SetRGB255(11, 79, 45)
	}
	tr.G.SetPixel(x - tr.xstart, y - tr.ystart)
	return
}

var width, height int = 256, 256

type TileRequest struct {
	G *gg.Context
	noise1, noise2 *Noise
	xstart, ystart, z int
}

func getTile(x, y, z int, seed1, seed2 string) *bytes.Buffer {
	xstart := width * x
	ystart := height * y

	noise1 := newNoise(seed1);
	noise2 := newNoise(seed2);

	g := gg.NewContext(width, height)

	tr := TileRequest{g, noise1, noise2, xstart, ystart, z}

	fmt.Printf("Tile Request for %v, %v, %v\n", z, x, y);
	//fmt.Printf("XStart: %v YStart: %v\n", xstart, ystart)

	for i := xstart; i < xstart + width; i++ {
		for j := ystart; j < ystart + height; j++ {
			e := getElevation(i, j, &tr)
			e_prev := getElevation(i-1, j-1, &tr)
			m := getMoisture(i, j, &tr)

			setTerrain(i, j, e, e_prev, m, &tr)
		}
	}

	out := new(bytes.Buffer)
	err := png.Encode(out, g.Image())
	if err != nil {
		fmt.Printf("ERROR: Cannot write image to PNG")
		//os.Exit(1)
	}

	//tr.G.SavePNG("output.png")

	return out
}

func TileHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "image/png")
	vars := mux.Vars(r)
	x, _ := strconv.Atoi(vars["x"])
	y, _ := strconv.Atoi(vars["y"])
	z, _ := strconv.Atoi(vars["z"])

	s1 := r.FormValue("s1")
	s2 := r.FormValue("s2")

	out := getTile(x, y, z, s1, s2)

	w.Write(out.Bytes())
}

var port = ":9797"

func main(){
	//seed1 := "seatSedaa"
	//seed2 := "dengor1"

	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET").Name("Tile").Path("/terraingo/tile/{z}/{x}/{y}.png").HandlerFunc(TileHandler)
	fmt.Println("API Listening")
	log.Fatal(http.ListenAndServe(port, router))

	//fmt.Println(createImage(widthin, heightin, seed1, seed2))
}