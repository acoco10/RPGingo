package main

import (
	"ShootEmUpAdventure/entities"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// structs
type Game struct {
	//game elements
	player       *entities.Player
	enemies      []*entities.Enemy
	items        []*entities.Item
	tilemapJSON  *TilemapJSON
	tilesets     []Tileset
	tilemapimage *ebiten.Image
	cam          *Camera
	colliders    []image.Rectangle
}

func CheckCollisionHorizontal(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		//check if player is colliding with collider
		if collider.Overlaps(
			image.Rect(
				int(sprite.X),
				int(sprite.Y),
				int(sprite.X)+16,
				int(sprite.Y)+16),
		) {
			if sprite.Dx > 0.0 { //check if player is going down
				//update player velocity
				sprite.X = float64(collider.Min.X) - 16.0
			} else if sprite.Dx < 0.0 { //check if player is going up
				sprite.X = float64(collider.Max.X)
			}

		}
	}
}

func CheckCollisionVertical(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		//check if player is colliding with collider
		if collider.Overlaps(
			image.Rect(
				int(sprite.X),
				int(sprite.Y),
				int(sprite.X)+16,
				int(sprite.Y)+16),
		) {
			if sprite.Dy > 0.0 { //check if player is going down
				//update player velocity
				sprite.Y = float64(collider.Min.Y) - 16.0
			} else if sprite.Dy < 0.0 { //check if player is going up
				sprite.Y = float64(collider.Max.Y)
			}

		}
	}
}

// game update function
func (g *Game) Update() error {

	g.player.Dx = 0
	g.player.Dy = 0
	//react to key presses by adding directional velocity
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.Dx = 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.Dx = -2
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Dy = 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Dy = -2
	}

	//increase players position by their velocity every update
	g.player.X += g.player.Dx

	CheckCollisionHorizontal(g.player.Sprite, g.colliders)

	g.player.Y += g.player.Dy

	CheckCollisionVertical(g.player.Sprite, g.colliders)

	for _, enemy := range g.enemies {

		enemy.Dx = 0.0

		enemy.Dy = 0.0

		if enemy.FollowsPlayer {
			if enemy.X < g.player.X {
				enemy.Dx += 1

			} else if enemy.X > g.player.X {
				enemy.Dx -= 1
			}
			if enemy.Y < g.player.Y {
				enemy.Dy += 1

			} else if enemy.Y > g.player.Y {
				enemy.Dy -= 1
			}
		}
		enemy.X += enemy.Dx
		CheckCollisionHorizontal(enemy.Sprite, g.colliders)
		enemy.Y += enemy.Dy
		CheckCollisionVertical(enemy.Sprite, g.colliders)
	}

	for _, item := range g.items {
		if math.Abs(item.X-g.player.X) <= 2 && math.Abs(item.Y-g.player.Y) <= 2 {
			g.player.Health += uint(item.AmtHeal)
			item.Ifinv = true
			fmt.Printf("Picked up an item! Health: %d\n", g.player.Health)
		}

	}

	g.cam.FollowTarget(g.player.X+16, g.player.Y+16, 320, 240)
	g.cam.Constrain(
		float64(g.tilemapJSON.Layers[0].Width)*16,
		float64(g.tilemapJSON.Layers[0].Height)*16,
		320,
		240,
	)

	return nil
}

// drawing screen + sprites
func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{120, 180, 255, 255})

	opts := ebiten.DrawImageOptions{}

	//map
	//loop through the tilemap
	for layerIndex, layer := range g.tilemapJSON.Layers {

		for index, id := range layer.Data {

			if id == 0 {
				continue
			}

			//coordinates example 1%30=1 1/30=0 2%30=2 2/30 = 0 etc...
			x := index % layer.Width
			y := index / layer.Width

			//pixel position
			x *= 16
			y *= 16

			img := g.tilesets[layerIndex].Img(id)

			opts.GeoM.Translate(float64(x), float64(y))

			opts.GeoM.Translate(0.0, -(float64(img.Bounds().Dy()) + 16))

			opts.GeoM.Translate(g.cam.X, g.cam.Y)

			screen.DrawImage(img, &opts)

			// reset the opts for the next tile
			opts.GeoM.Reset()

		}
	}

	//draw player

	opts.GeoM.Translate(g.player.X, g.player.Y)
	opts.GeoM.Translate(g.cam.X, g.cam.Y)

	screen.DrawImage(
		//grab a subimage of the Spritesheet
		g.player.Img.SubImage(
			image.Rect(0, 0, 16, 32),
		).(*ebiten.Image),
		&opts,
	)

	opts.GeoM.Reset()

	// draw enemy sprites
	for _, sprite := range g.enemies {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)

		screen.DrawImage(
			sprite.Img.SubImage(
				image.Rect(0, 0, 16, 16),
			).(*ebiten.Image),
			&opts,
		)

		opts.GeoM.Reset()

	}

	//draw item
	for _, sprite := range g.items {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)
		if !sprite.Ifinv {
			screen.DrawImage(
				sprite.Img.SubImage(
					image.Rect(0, 0, 16, 16),
				).(*ebiten.Image),
				&opts,
			)
		}

		opts.GeoM.Reset()
	}

	for _, collider := range g.colliders {
		vector.StrokeRect(
			screen,
			float32(collider.Min.X)+float32(g.cam.X),
			float32(collider.Min.Y)+float32(g.cam.Y),
			float32(collider.Dx()),
			float32(collider.Dy()),
			1.0,
			color.RGBA{255, 0, 0, 255},
			false,
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ebiten.WindowSize()
}

func main() {
	ebiten.SetWindowSize(320, 240)
	ebiten.SetWindowTitle("Quick Draw Adventure")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/characters/Elysefront.png")
	if err != nil {
		//handle error
		log.Fatal(err)
	}

	ghostImg, _, err := ebitenutil.NewImageFromFile("assets/images//enemies/Ghost.png")
	if err != nil {
		//handle error
		log.Fatal(err)
	}

	fishImg, _, err := ebitenutil.NewImageFromFile("assets/images//items/Fish.png")
	if err != nil {
		//handle error
		log.Fatal(err)
	}

	tilemapimage, _, err := ebitenutil.NewImageFromFile("assets/images/terrain/TilesetFloor.png")
	if err != nil {
		//handle error
		log.Fatal(err)
	}

	tilemapJSON, err := NewTilemapJSON("assets/map/demoMap.json")
	if err != nil {
		//handle error
		log.Fatal(err)
	}
	tilesets, err := tilemapJSON.GenTileSets()

	if err != nil {

		log.Fatalf("Failed to generate tilesets: %v", err)
	}

	game := Game{
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img: playerImg,
				X:   75.0,
				Y:   75.0,
			},
			Health: 10,
		},
		enemies: []*entities.Enemy{
			{
				Sprite: &entities.Sprite{
					Img: ghostImg,
					X:   100.0,
					Y:   100.0,
				},
				FollowsPlayer: true,
			},
			{
				Sprite: &entities.Sprite{
					Img: ghostImg,
					X:   50.0,
					Y:   50.0,
				},
				FollowsPlayer: false,
			},
			{
				Sprite: &entities.Sprite{
					Img: ghostImg,
					X:   100.0,
					Y:   100.0,
				},
				FollowsPlayer: false,
			},
		},
		items: []*entities.Item{
			{
				Sprite: &entities.Sprite{
					Img: fishImg,
					X:   335.0,
					Y:   335.0,
				},
				AmtHeal: rand.Intn(4),
				Ifinv:   false,
			},
		},
		tilemapJSON:  tilemapJSON,
		tilemapimage: tilemapimage,
		tilesets:     tilesets,
		cam:          NewCamera(0.0, 0.0),
		colliders: []image.Rectangle{
			image.Rect(100, 100, 116, 116),
		},
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
