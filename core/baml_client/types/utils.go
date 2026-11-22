package types
import (
	baml "github.com/boundaryml/baml/engine/language_client_go/pkg"
)
type Checked[T any] = baml.Checked[T]
type Image baml.Image
type Audio baml.Audio
type Video baml.Video
type PDF baml.PDF