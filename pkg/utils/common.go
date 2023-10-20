package utils

const (
	// WorkspaceFinalizer is used to make sure that workspace controller handles garbage collection.
	WorkspaceFinalizer = "workspace.finalizer.kaito.sh"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
