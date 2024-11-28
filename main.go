package main

import (
	"flag"
	"fmt"
	"hash/fnv"
	"image/color"
	"log"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const sv_friction float32 = 6
const sv_stopspeed float32 = 1

const MAX_SPEED float32 = 4
const MAX_AIR_SPEED float32 = 0.5

const Pi = math.Pi

const sv_accelerate = 16
const gravity float32 = -16

var hasher = fnv.New64a()

func getWishDir(cameraDirection rl.Vector3) rl.Vector3 {
	var dx float64 = 0
	var dy float64 = 0

	if rl.IsKeyDown(rl.KeyW) {
		dx += 1
	}
	if rl.IsKeyDown(rl.KeyS) {
		dx -= 1
	}

	if rl.IsKeyDown(rl.KeyD) {
		dy += 1
	}
	if rl.IsKeyDown(rl.KeyA) {
		dy -= 1
	}

	var angle float32
	if dx == 1 {
		switch dy {
		case 1:
			angle = Pi / 4
		case 0:
			angle = 0
		case -1:
			angle = -Pi / 4
		}
	} else if dx == -1 {
		switch dy {
		case 1:
			angle = 3 * Pi / 4
		case -0:
			angle = Pi
		case -1:
			angle = -3 * Pi / 4
		}

	} else {
		switch dy {
		case 1:
			angle = Pi / 2
		case 0:
			return rl.Vector3{0, 0, 0}
		case -1:
			angle = -Pi / 2

		}
	}

	rot := rl.Vector2Rotate(rl.Vector2{cameraDirection.X, cameraDirection.Z}, angle)
	// return rl.Vector3Normalize(rl.Vector3{dx, dy, 0})
	return rl.Vector3Normalize(rl.Vector3{rot.X, rot.Y, 0})
}

type Cube struct {
	position rl.Vector3
	width    float32
	height   float32
	length   float32
	col      color.RGBA
}

var world []Cube = []Cube{
	Cube{rl.Vector3{0, 0, 0}, 10, 0.1, 10, rl.DarkGray},
	Cube{rl.Vector3{1, 0, 0}, 1, 1, 1, rl.Blue},
	Cube{rl.Vector3{3, 0.2, 2}, 1, 1, 1, rl.Green},
	Cube{rl.Vector3{5, 0.4, 3}, 1, 1, 1, rl.Orange},
	Cube{rl.Vector3{7, 0.2, 1}, 1, 1, 1, rl.Purple}}

func (x Cube) render() {
	rl.DrawCube(x.position, x.width, x.height, x.length, x.col)
}

func drawMap() {
	for _, obj := range world {
		obj.render()
	}
}

func friction(vel *rl.Vector3, frametime float32) {
	speed := float32(math.Sqrt(float64(vel.X*vel.X + vel.Y*vel.Y)))
	if speed == 0 {
		return
	}

	friction := sv_friction

	// apply friction
	var control float32
	if speed < sv_stopspeed {
		control = sv_stopspeed
	} else {
		control = speed
	}

	var newspeed float32 = speed - (frametime * control * friction)

	if newspeed < 0 {
		newspeed = 0
	}
	newspeed /= speed
	(*vel).X *= newspeed
	(*vel).Y *= newspeed
	(*vel).Z *= newspeed
}

func accelerate(wishspeed float32, velocity *rl.Vector3, wishdir rl.Vector3, frametime float32) {
	currentspeed := rl.Vector3DotProduct(*velocity, wishdir)
	addspeed := wishspeed - currentspeed
	if addspeed <= 0 {
		return
	}

	accelspeed := sv_accelerate * frametime * wishspeed
	if accelspeed > addspeed {
		accelspeed = addspeed
	}

	(*velocity).X += wishdir.X * accelspeed
	(*velocity).Y += wishdir.Y * accelspeed
	(*velocity).Z += wishdir.Z * accelspeed
}

func airAccelerate(wishspeed float32, velocity *rl.Vector3, wishdir rl.Vector3, frametime float32) {
	wishspd := float32(math.Sqrt(float64(rl.Vector3DotProduct(wishdir, wishdir))))
	wishdir = rl.Vector3Normalize(wishdir)

	if wishspd > MAX_AIR_SPEED {
		wishspd = MAX_AIR_SPEED
	}

	currentspeed := rl.Vector3DotProduct(*velocity, wishdir)
	addspeed := wishspd - currentspeed
	if addspeed <= 0 {
		return
	}

	accelspeed := sv_accelerate * wishspeed * frametime
	if accelspeed > addspeed {
		accelspeed = addspeed
	}

	(*velocity).X += wishdir.X * accelspeed
	(*velocity).Y += wishdir.Y * accelspeed
	(*velocity).Z += wishdir.Z * accelspeed
}

func onGround(pos rl.Vector3, size rl.Vector3) bool {
	entityBox := rl.BoundingBox{rl.Vector3{pos.X - size.X/2, pos.Y - size.Y/2, pos.Z - size.Z/2},
		rl.Vector3{pos.X + size.X/2, pos.Y + size.Y/2, pos.Z + size.Z/2}}

	for _, obj := range world {
		//col := GetRayCollisionBox(rl.Ray{pos, rl.Vector3{0, -1, 0}}, box BoundingBox) RayCollision
		objBox := rl.BoundingBox{rl.Vector3{obj.position.X - obj.width/2, obj.position.Y - obj.height/2, obj.position.Z - obj.length/2},
			rl.Vector3{obj.position.X + obj.width/2, obj.position.Y + obj.height/2, obj.position.Z + obj.length/2}}

		//rl.DrawBoundingBox(objBox, rl.Green);

		if rl.CheckCollisionBoxes(entityBox, objBox) && (pos.Y-size.Y/2 <= obj.position.Y+obj.height/2) {
			return true
		}
	}
	return false
}

func limitPitchAngle(angle float32, up rl.Vector3, targetPos rl.Vector3) float32 {
	// Clamp view up
	maxAngleUp := rl.Vector3Angle(up, targetPos)
	maxAngleUp = maxAngleUp - 0.01 // avoid numerical errors
	if angle > maxAngleUp {
		angle = maxAngleUp
	}

	// Clamp view down
	maxAngleDown := rl.Vector3Angle(rl.Vector3Negate(up), targetPos)
	maxAngleDown = -maxAngleDown + 0.01 // avoid numerical errors
	if angle < maxAngleDown {
		angle = maxAngleDown
	}
	return angle
}

type Player struct {
	Position rl.Vector3
	Velocity rl.Vector3
	Size     rl.Vector3
}

func hashString(inp string) uint64 {
	hasher.Write([]byte(inp))
	s := hasher.Sum64()
	hasher.Reset()
	return s
}

func main() {
	seedf := flag.String("seed", "", "Seed of the world to generate. No seed/empty seed will generate a random map.")
	flag.Parse()

	rl.InitWindow(1920, 1080, "stanleymw's movement test")
	defer rl.CloseWindow()
	rl.SetTargetFPS(240)
	rl.DisableCursor()


	var camera = rl.NewCamera3D(rl.Vector3{0, 0, 0}, rl.Vector3{1, 0, 0}, rl.Vector3{0, 1, 0}, 90.0, rl.CameraPerspective)
	var player Player = Player{rl.Vector3{0, 5, 0}, rl.Vector3{0, 0, 0}, rl.Vector3{1, 2, 1}}
	var targetPosition = rl.Vector3{1, 0, 0}

	var seed int64
	if (*seedf == "") {
		seed = rand.Int63()
	} else {
		seed = int64(hashString(*seedf))
	}
	log.Printf("Generating world with seed: %d from user: %s", seed, *seedf)
	gen := rand.New(rand.NewSource(seed))

	lastGeneratedPos := rl.Vector3{0, 0, 0}
	for i := 0; i < 5e3; i++ {
		dx := gen.Float32()*10.0 - 5.0
		dy := gen.Float32()*2.0 - 1.0
		dz := gen.Float32()*10.0 - 5.0
		lastGeneratedPos = rl.Vector3Add(lastGeneratedPos, rl.Vector3{dx, dy, dz})
		// pos := rl.Vector3{X: float32(10 * math.Cos(float64(i)/20.0*math.Pi)), Y: float32(i) / 20.0, Z: float32(10 * math.Sin(float64(i)/20.0*math.Pi))}
		color := color.RGBA{uint8(i%120 + 50), 0, 0, 255}

		world = append(world, Cube{lastGeneratedPos, 1.5, 0.2, 1.5, color})
	}

	world = append(world, Cube{lastGeneratedPos, 16, 0.4, 16, color.RGBA{20, 170, 195, 255}})

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		var frametime = rl.GetFrameTime()
		cameraRotation := rl.GetMouseDelta()

		// apply X rotation
		targetPosition = rl.Vector3RotateByAxisAngle(targetPosition, rl.GetCameraUp(&camera), -cameraRotation.X*0.0012)
		// apply Y rotation
		targetPosition = rl.Vector3RotateByAxisAngle(targetPosition, rl.GetCameraRight(&camera), limitPitchAngle(-cameraRotation.Y*0.0012, rl.GetCameraUp(&camera), targetPosition))

		var wishdir = getWishDir(targetPosition)

		// update player Velocity
		var grounded = onGround(player.Position, player.Size)

		if grounded {
			// on ground
			if rl.IsKeyDown(rl.KeySpace) || rl.GetMouseWheelMove() != 0 {
				player.Velocity.Z = 6
			} else {
				player.Velocity.Z = 0
				friction(&player.Velocity, frametime)
				accelerate(MAX_SPEED, &player.Velocity, wishdir, frametime)
			}
		} else {
			// in air
			airAccelerate(MAX_SPEED, &player.Velocity, wishdir, frametime)
			player.Velocity.Z += gravity * frametime
		}

		// update player position
		player.Position.X += player.Velocity.X * frametime
		player.Position.Y += player.Velocity.Z * frametime
		player.Position.Z += player.Velocity.Y * frametime


		if rl.IsKeyDown(rl.KeyR) {
			player.Position = rl.Vector3{0, 3, 0}
			player.Velocity = rl.Vector3{0, 0, 0}
		}

		// update camera position
		camera.Position.X = player.Position.X
		camera.Position.Y = player.Position.Y
		camera.Position.Z = player.Position.Z

		// set camera rotation
		camera.Target = rl.Vector3Add(camera.Position, targetPosition)

		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode3D(camera)
		drawMap()
		rl.EndMode3D()

		rl.DrawText(fmt.Sprintf(" %d fps\n %.2f, %.2f, %.2f\n vel=%.2f\n r: %.2f, %.2f\n %f %f\n cam=%v\n wd=%v\n g=%t \n mw=%f",
			rl.GetFPS(),
			player.Position.X,
			player.Position.Y,
			player.Position.Z,
			rl.Vector2Length(rl.Vector2{X: player.Velocity.X, Y: player.Velocity.Y}),
			cameraRotation.X,
			cameraRotation.Y,
			rl.Vector3Angle(rl.GetCameraUp(&camera), targetPosition),
			rl.Vector3Angle(rl.GetCameraForward(&camera), targetPosition),
			targetPosition,
			wishdir,
			grounded,
			rl.GetMouseWheelMove()), 0, 0, 32, rl.Black)

		rl.EndDrawing()
	}
}
