package utils

import (
	"fmt"
	"github.com/frajibe/piped-playfeed/settings"
	"github.com/schollz/progressbar/v3"
	"os"
)

func CreateProgressBar(max int, description string) *progressbar.ProgressBar {
	if settings.GetSettingsService().SilentMode {
		return nil
	}
	max = adjustMaxValue(max)
	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
	)
}

func IncrementProgressBar(progressBar *progressbar.ProgressBar) {
	if progressBar != nil {
		progressBar.Add(1)
	}
}

func CreateInfiniteProgressBar(description string) *progressbar.ProgressBar {
	return CreateProgressBar(1, description)
}

func FinalizeProgressBar(progressBar *progressbar.ProgressBar, max int) {
	if progressBar != nil {
		max = adjustMaxValue(max)
		progressBar.ChangeMax(max)
		progressBar.Set(max)
		progressBar.Finish()
	}
}

func adjustMaxValue(max int) int {
	if max == 0 {
		return 1
	}
	return max
}
