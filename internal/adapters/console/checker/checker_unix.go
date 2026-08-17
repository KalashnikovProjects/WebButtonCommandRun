//go:build !windows

package checker

type Checker struct{}

func New() *Checker {
	return &Checker{}
}

// On unix always available
func (checker *Checker) CheckAvailability() error {
	return nil
}
