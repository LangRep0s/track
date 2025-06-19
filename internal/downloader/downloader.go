package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)


func DownloadFile(url, filepath string) error {
	
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(180*time.Millisecond),
	)

	bar := p.New(resp.ContentLength,
		mpb.BarStyle().Lbound("[").Filler("=").Tip("> ").Padding("-").Rbound("]"),
		mpb.PrependDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.Name(" ] "),
			decor.EwmaSpeed("KiB", "% .2f", 60),
		),
	)

	
	proxyReader := bar.ProxyReader(resp.Body)
	defer proxyReader.Close()

	
	_, err = io.Copy(out, proxyReader)
	if err != nil {
		return err
	}

	p.Wait()

	return nil
}
