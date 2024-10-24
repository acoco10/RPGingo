package main

import (
	"ShootEmUpAdventure/animations"
	"ShootEmUpAdventure/entities"
	"ShootEmUpAdventure/mapobjects"
	"ShootEmUpAdventure/spritesheet"
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
	player            *entities.Player
	playerSpriteSheet *spritesheet.SpriteSheet
	enemies           []*entities.Enemy
	items             []*entities.Item
	tilemapJSON       *mapobjects.TilemapJSON
	tilesets          []mapobjects.Tileset
	cam               *Camera
	colliders         []image.Rectangle
	entranceDoors     map[string]mapobjects.Door
	exitDoors         map[string]mapobjects.Door
}

// game update function
func (g *Game) Update() error {

	if len(g.entranceDoors) == 0 {
		//loading objects(colliders, doors)from tilesetJSON data
		mapobjects.StoreMapObjects(*g.tilemapJSON, &g.colliders, g.entranceDoors, g.exitDoors)
	}
	g.player.Dx = 0
	g.player.Dy = 0

	//react to key presses by adding directional velocity
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.Dx = 1.5
		g.player.Direction = "L"
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.Dx = -1.5
		g.player.Direction = "R"
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Dy = 1.5
		g.player.Direction = "U"
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Dy = -1.5
		g.player.Direction = "D"
	}

	//increase players position by their velocity every update
	g.player.X += g.player.Dx

	mapobjects.CheckCollisionHorizontal(g.player.Sprite, g.colliders)

	g.player.Y += g.player.Dy

	mapobjects.CheckCollisionVertical(g.player.Sprite, g.colliders)

	mapobjects.CheckEnterDoor(g.player, g.entranceDoors, g.exitDoors)

	mapobjects.CheckExitDoor(g.player, g.entranceDoors, g.exitDoors)

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

		mapobjects.CheckCollisionHorizontal(enemy.Sprite, g.colliders)

		enemy.Y += enemy.Dy

		mapobjects.CheckCollisionVertical(enemy.Sprite, g.colliders)

	}

	activAnimation := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
	if activAnimation != nil {
		activAnimation.Update()
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

	opts := ebiten.DrawImageOptions{}

	//map
	//loop through the tilemap

	for _, layer := range g.tilemapJSON.Layers {
		if layer.Type == "objectgroup" {
			continue
		}

		for index, id := range layer.Data {

			if id == 0 {
				continue
			}

			tileindex := 0

			for i := range len(g.tilesets) - 1 {
				if id < g.tilesets[i].Gid() {
					tileindex -= 1
				}
				if id >= g.tilesets[i+1].Gid() {
					tileindex += 1
				}
			}

			//coordinates example 1%30=1 1/30=0 2%30=2 2/30 = 0 etc...

			x := index % layer.Width
			y := index / layer.Width

			//pixel position
			x *= 16
			y *= 16

			img := g.tilesets[tileindex].Img(id)

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

	playerFrame := 0
	activAnimation := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
	if activAnimation != nil {

		playerFrame = activAnimation.Frame()
	} else {
		if g.player.Direction == "U" {
			playerFrame = g.player.Animations[0].FirstF
		}
		if g.player.Direction == "D" {
			playerFrame = g.player.Animations[1].FirstF
		}
		if g.player.Direction == "R" {
			playerFrame = g.player.Animations[2].FirstF
		}
		if g.player.Direction == "L" {
			playerFrame = g.player.Animations[3].FirstF
		}

	}

	screen.DrawImage(
		//grab a subimage of the Spritesheet
		g.player.Img.SubImage(
			g.playerSpriteSheet.Rect(playerFrame),
		).(*ebiten.Image),
		&opts,
	)

	opts.GeoM.Reset()

	for _, layer := range g.tilemapJSON.Layers {
		if layer.Class != "top" {
			continue
		}

		for index, id := range layer.Data {

			if id == 0 {
				continue
			}

			tileindex := 0

			for i := range len(g.tilesets) - 1 {
				if id < g.tilesets[i].Gid() {
					tileindex -= 1
				}
				if id >= g.tilesets[i+1].Gid() {
					tileindex += 1
				}
			}

			//coordinates example 1%30=1 1/30=0 2%30=2 2/30 = 0 etc...

			x := index % layer.Width
			y := index / layer.Width

			//pixel position
			x *= 16
			y *= 16

			if int(g.player.Y)+48 < y {

				img := g.tilesets[tileindex].Img(id)

				opts.GeoM.Translate(float64(x), float64(y))

				opts.GeoM.Translate(0.0, -(float64(img.Bounds().Dy()) + 16))

				opts.GeoM.Translate(g.cam.X, g.cam.Y)

				screen.DrawImage(img, &opts)

				// reset the opts for the next tile
				opts.GeoM.Reset()
			}

		}
	}

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

	//TESTING drawing colliders for testing

	vector.StrokeRect(
		screen,
		float32(g.player.X)+float32(g.cam.X),
		float32(g.player.Y)+27+float32(g.cam.Y),
		14,
		4,
		1.0,
		color.RGBA{255, 0, 0, 255},
		false,
	)

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

	//drawing doors for testing
	for _, door := range g.entranceDoors {
		vector.StrokeRect(
			screen,
			float32(door.Coord.Min.X)+float32(g.cam.X),
			float32(door.Coord.Min.Y)+float32(g.cam.Y),
			float32(door.Coord.Dx()),
			float32(door.Coord.Dy()),
			1.0,
			color.RGBA{255, 0, 0, 255},
			false,
		)
	}

	for _, door := range g.exitDoors {
		vector.StrokeRect(
			screen,
			float32(door.Coord.Min.X)+float32(g.cam.X),
			float32(door.Coord.Min.Y)+float32(g.cam.Y),
			float32(door.Coord.Dx()),
			float32(door.Coord.Dy()),
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

	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/characters/elyseSpriteSheet.png")
	if err != nil {
		//handle error
		log.Fatal(err)
	}
	/*
		ghostImg, _, err := ebitenutil.NewImageFromFile("assets/images//enemies/Ghost.png")
		if err != nil {
			//handle error
			log.Fatal(err)
		} */

	fishImg, _, err := ebitenutil.NewImageFromFile("assets/images//items/Fish.png")
	if err != nil {
		//handle error
		log.Fatal(err)
	}

	tilemapJSON, err := mapobjects.NewTilemapJSON("assets/map/town1Map.json")

	if err != nil {
		//handle error
		log.Fatal(err)
	}

	tilesets, err := tilemapJSON.GenTileSets()

	if err != nil {
		log.Fatalf("Failed to generate tilesets: %v", err)
	}

	playerSpriteSheet := spritesheet.NewSpritesheet(4, 4, 18, 18, 31)

	//player running animation calling animation package new animation function

	game := Game{
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img: playerImg,
				X:   125,
				Y:   125,
			},
			Health: 10,
			Animations: map[entities.PlayerState]*animations.Animation{
				entities.Down:  animations.NewAnimation(0, 4, 4, 22.0),
				entities.Up:    animations.NewAnimation(2, 6, 4, 22.0),
				entities.Left:  animations.NewAnimation(1, 10, 4, 11.0),
				entities.Right: animations.NewAnimation(3, 11, 4, 11.0),
			},
		},

		playerSpriteSheet: playerSpriteSheet,
		/* 	enemies: []*entities.Enemy{ */
		/* {
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
		}, */
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
		tilemapJSON:   tilemapJSON,
		tilesets:      tilesets,
		cam:           NewCamera(0.0, 0.0),
		entranceDoors: make(map[string]mapobjects.Door),
		exitDoors:     make(map[string]mapobjects.Door),
		colliders:     make([]image.Rectangle, 0),
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
