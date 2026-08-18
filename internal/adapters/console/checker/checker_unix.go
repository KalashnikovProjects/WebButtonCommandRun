//go:build !windows

package checker

type Checker struct {
	ptyDir string
}

func New(ptyDir string) *Checker {
	return &Checker{
		ptyDir: ptyDir,
	}
}

// On unix always available
func (ch *Checker) CheckAvailability() error {
	return nil
}
