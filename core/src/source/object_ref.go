package source
import (
	"encoding/base64"
	"encoding/json"
	"slices"
)
type objectRefLocatorPayload struct {
	Kind ObjectKind `json:"kind"`
	Path []string   `json:"path"`
}
func NewObjectRef(kind ObjectKind, path []string) ObjectRef {
	ref := ObjectRef{
		Kind: kind,
		Path: slices.Clone(path),
	}
	ref.Locator = EncodeObjectRefLocator(ref)
	return ref
}
func EncodeObjectRefLocator(ref ObjectRef) string {
	payload, err := json.Marshal(objectRefLocatorPayload{
		Kind: ref.Kind,
		Path: slices.Clone(ref.Path),
	})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(payload)
}
func DecodeObjectRefLocator(locator string) (ObjectRef, error) {
	if locator == "" {
		return ObjectRef{}, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(locator)
	if err != nil {
		return ObjectRef{}, err
	}
	var payload objectRefLocatorPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ObjectRef{}, err
	}
	return ObjectRef{
		Kind:    payload.Kind,
		Path:    slices.Clone(payload.Path),
		Locator: locator,
	}, nil
}
func NormalizeObjectRef(ref ObjectRef) ObjectRef {
	normalized := ObjectRef{
		Kind:    ref.Kind,
		Path:    slices.Clone(ref.Path),
		Locator: ref.Locator,
	}
	if len(normalized.Path) == 0 && normalized.Locator != "" {
		if decoded, err := DecodeObjectRefLocator(normalized.Locator); err == nil {
			if normalized.Kind == "" {
				normalized.Kind = decoded.Kind
			}
			normalized.Path = decoded.Path
		}
	}
	if normalized.Locator == "" {
		normalized.Locator = EncodeObjectRefLocator(normalized)
	}
	return normalized
}