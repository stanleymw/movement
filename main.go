package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"fmt"
	"math"
	"image/color"
)

const sv_friction float32 = 7
const sv_stopspeed float32 = 1

const MAX_SPEED float32 = 4
const MAX_AIR_SPEED float32 = 0.375

const sv_accelerate = 10

func getWishDir() rl.Vector3 {
	var dx float32
	var dy float32
	if (rl.IsKeyDown(rl.KeyW)) {
		dx += 1
	}
	if (rl.IsKeyDown(rl.KeyS)) {
		dx -= 1
	}

	if (rl.IsKeyDown(rl.KeyD)) {
		dy += 1
	}
	if (rl.IsKeyDown(rl.KeyA)) {
		dy -= 1
	}

	return rl.Vector3Normalize(rl.Vector3{dx, dy, 0})
}

type Cube struct {
	position rl.Vector3
	width float32
	height float32
	length float32
	col color.RGBA
}

var world []Cube = []Cube{
	Cube{rl.Vector3{0,0,0}, 10, 0.1, 10, rl.Red},
	Cube{rl.Vector3{1,0,0}, 1, 1, 1, rl.Blue},
	Cube{rl.Vector3{3,0.2,2}, 1, 1, 1, rl.Green},
	Cube{rl.Vector3{5,0.4,3}, 1, 1, 1, rl.Orange},
	Cube{rl.Vector3{7,0.2,1}, 1, 1, 1, rl.Purple}}

func (x Cube) render() {
	rl.DrawCube(x.position, x.width, x.height, x.length, x.col)
}

func drawMap() {
	for _, obj := range world {
		obj.render()
	}
}

func friction(vel *rl.Vector3, frametime float32) {
	speed := float32(math.Sqrt(float64(vel.X * vel.X + vel.Y * vel.Y)))
	if (speed == 0) {
		return
	}

	friction := sv_friction

	// apply friction
	var control float32
	if (speed < sv_stopspeed) {
		control = sv_stopspeed
	} else {
		control = speed
	}

	var newspeed float32 = speed - (frametime * control * friction)

	if (newspeed < 0) {
		newspeed = 0
	}
	newspeed /= speed
	(*vel).X *= newspeed
	(*vel).Y *= newspeed
	(*vel).Z *= newspeed
}

func accelerate(wishspeed float32, velocity *rl.Vector3, wishdir rl.Vector3, frametime float32) {
	currentspeed := rl.Vector3DotProduct(*velocity, wishdir);
	addspeed := wishspeed - currentspeed;
	if (addspeed <= 0) {
		return
	}

	accelspeed := sv_accelerate * frametime * wishspeed;
	if (accelspeed > addspeed) {
		accelspeed = addspeed;
	}

	(*velocity).X += wishdir.X * accelspeed
	(*velocity).Y += wishdir.Y * accelspeed
	(*velocity).Z += wishdir.Z * accelspeed
}

func airAccelerate(wishspeed float32, velocity *rl.Vector3, wishdir rl.Vector3, frametime float32) {
	wishspd := float32(math.Sqrt(float64(rl.Vector3DotProduct(wishdir, wishdir))));
	wishdir = rl.Vector3Normalize(wishdir)

	if (wishspd > MAX_AIR_SPEED) {
		wishspd = MAX_AIR_SPEED;
	}

	currentspeed := rl.Vector3DotProduct(*velocity, wishdir);
	addspeed := wishspd - currentspeed;
	if (addspeed <= 0) {
		return;
	}

	accelspeed := sv_accelerate * wishspeed * frametime;
	if (accelspeed > addspeed) {
		accelspeed = addspeed;
	}

	(*velocity).X += wishdir.X * accelspeed
	(*velocity).Y += wishdir.Y * accelspeed
	(*velocity).Z += wishdir.Z * accelspeed
}


func onGround(pos rl.Vector3, size rl.Vector3) bool {
	entityBox := rl.BoundingBox{rl.Vector3{pos.X - size.X/2, pos.Y - size.Y/2, pos.Z - size.Z/2 },
				rl.Vector3{pos.X + size.X/2, pos.Y + size.Y/2, pos.Z + size.Z/2 }};

	for _, obj := range world {
		//col := GetRayCollisionBox(rl.Ray{pos, rl.Vector3{0, -1, 0}}, box BoundingBox) RayCollision
		objBox := rl.BoundingBox{rl.Vector3{obj.position.X - obj.width/2, obj.position.Y - obj.height/2, obj.position.Z - obj.length/2 },
					rl.Vector3{obj.position.X + obj.width/2, obj.position.Y + obj.height/2, obj.position.Z + obj.length/2 }};

		//rl.DrawBoundingBox(objBox, rl.Green);

		if (rl.CheckCollisionBoxes(entityBox, objBox) && (pos.Y - size.Y/2 <= obj.position.Y + obj.height / 2)) {
			return true
		}
	}
	return false
}

func main() {
	rl.InitWindow(1920, 1080, "stanleymw's movement test")
	defer rl.CloseWindow()

	rl.SetTargetFPS(240)

	var camera = rl.NewCamera3D(rl.Vector3{0,5,0}, rl.Vector3{1,0,0}, rl.Vector3{0,1,0}, 90.0, rl.CameraPerspective)
	
	rl.DisableCursor()

	var gravity float32 = -16
	var velocity = rl.Vector3{0, 0, 0}

	var playerSize = rl.Vector3{1, 2, 1}

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		var frametime = rl.GetFrameTime()
		//var wishdir = getWishDir()
		var wishdir = getWishDir()
		var grounded = onGround(camera.Position, playerSize)

		if grounded {
			// on ground
			if (rl.IsKeyDown(rl.KeySpace)) {
				velocity.Z = 6
			} else {
				velocity.Z = 0
				friction(&velocity, frametime)
				accelerate(MAX_SPEED, &velocity, wishdir, frametime)
			}
		} else {
			// in air
			airAccelerate(MAX_SPEED, &velocity, wishdir, frametime)
			velocity.Z += gravity * frametime
		}
		// velocity = rl.Vector3{wishdir.X, wishdir.Y, velocity.Z}

		rl.UpdateCameraPro(&camera,
			rl.Vector3Scale(velocity, frametime),
			rl.Vector3{rl.GetMouseDelta().X * 0.05, rl.GetMouseDelta().Y * 0.05, 0.0},
			0.0)

		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode3D(camera)
			drawMap()
		rl.EndMode3D()

		rl.DrawText(fmt.Sprintf(" %d fps\n %.3f, %.3f, %.3f\n %f\n %s", 
		rl.GetFPS(), 
		camera.Position.X, 
		camera.Position.Y, 
		camera.Position.Z,
		rl.Vector2Length(rl.Vector2{velocity.X, velocity.Y}),
		grounded), 0, 0, 32, rl.Black)

		rl.EndDrawing()
	}
}
