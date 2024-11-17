package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"fmt"
)

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

func main() {
	rl.InitWindow(1920, 1080, "stanleymw's movement test")
	defer rl.CloseWindow()

	rl.SetTargetFPS(240)

	var camera = rl.NewCamera3D(rl.Vector3{0,5,0}, rl.Vector3{1,0,0}, rl.Vector3{0,1,0}, 90.0, rl.CameraPerspective)
	
	rl.DisableCursor()

	var _ = rl.Vector3{0, 0, -1}
	var velocity = rl.Vector3{0, 0, 0}

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		var frametime = rl.GetFrameTime()

		//var wishdir = getWishDir()
		var wishdir = rl.Vector3Scale(getWishDir(), frametime)
		velocity = rl.Vector3{wishdir.X, wishdir.Y, velocity.Z}

		rl.UpdateCameraPro(&camera,
			velocity,
			rl.Vector3{rl.GetMouseDelta().X * 0.05, rl.GetMouseDelta().Y * 0.05, 0.0},
			0.0)

		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode3D(camera)
			rl.DrawCube(rl.Vector3{0,0,0}, 5, 0.1, 5, rl.Red)
		rl.EndMode3D()

		rl.DrawText(fmt.Sprintf(" %d fps\n %.3f, %.3f, %.3f", rl.GetFPS(), camera.Position.X, camera.Position.Y, camera.Position.Z), 0, 0, 32, rl.Black)

		rl.EndDrawing()
	}
}
