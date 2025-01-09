package videodownloader

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const videoPathDirectory = "video"
const audioPathDirectory = "audio"
const defaultVideoQuality = 700
const MaxDuration = time.Second*60*30
const videoPrefix = "video_"
const audioPrefix = "audio_"
const outputPathDirectory = "output"
const TooLong = "1"
const TooBig = 50000000
const FileIsTooBig = "2"
// Download downloads a YouTube video given its URL.
//
// Download is a convenience function that parses a YouTube video URL and
// downloads the video.

func Download(url string) (string, string, error) {
	videoID, err := getVideoIDfromURL(url)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	//check if file already downloaded
	if _, err := os.Stat(outputPathDirectory+ "/" + videoID + ".mp4"); err == nil {
	   fmt.Println("!already have in cache")
	   return (outputPathDirectory+ "/" + videoID + ".mp4"), "cachedVideo", nil
	} 

	videoFileName, requiresAudio, title, err := downloadVideo(videoID)

	if !requiresAudio {
		return videoFileName, title, nil
	}

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	
	//if audio is present skip merge
	

	audioFileName, err := downloadAudio(videoID)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	mergedVideName, err  := mergeVideoAndAudio(videoFileName, audioFileName)
	if err != nil {
		fmt.Println(err)
		return "", "",err
	}

	deleteVideAndAudio()
	
	return mergedVideName, title,  nil
}


func getVideoIDfromURL(youtubeurl string) (string, error) {
	
	

	u, err := url.Parse(youtubeurl)
	if err != nil {
		return "", err
	}



	//fmt.Println(u.Scheme)
	if u.Scheme != "https" {
		return "", fmt.Errorf("Invalid URL scheme: %s", u.Scheme, youtubeurl)
	}

	if _, ok := u.Query()["si"]; ok {
		
		return u.Path[1:], nil
	}

	if _, ok :=  u.Query()["v"]; ok {
		return u.Query()["v"][0], nil
	}

	return "", fmt.Errorf("invalid URL query: %s", u.Query())
}

// downloadVideo downloads the video content given a YouTube video ID.
func downloadVideo (id string) (string, bool, string, error) {
	
	requiresAudio := true

	videoID := id
	client := youtube.Client{}
	video, err := client.GetVideo(videoID)
	if err != nil {
		fmt.Println(err)
		return "", false, "", err
	}

	formats := video.Formats
	
	fmtIndex := 0
	for i, f := range formats {
		fmt.Println(f.ApproxDurationMs, f.AudioChannels, f.Quality, f.Width, f.ApproxDurationMs)
		if f.Width < defaultVideoQuality {
			
			fmtIndex = i
			if f.AudioChannels > 0 {
				requiresAudio = false
			}
			break		
		}
	}
	fmt.Println(video.Duration, MaxDuration)
	if video.Duration > MaxDuration {
		return TooLong, false, "", fmt.Errorf("Video is too long: %v", video.Duration)
	}

	stream, _, err := client.GetStream(video, &formats[fmtIndex])
	if err != nil {
		fmt.Println(err)
		return "", false, "", err
	}
	defer stream.Close()

	fileName := videoPathDirectory + "/" + videoPrefix + videoID +".mp4"
	fmt.Print(fileName)

	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		panic(err)
	}

	


	return fileName, requiresAudio, video.Title, nil 
}

func downloadAudio (id string) (string, error) {
	

	videoID := id
	client := youtube.Client{}
	video, err := client.GetVideo(videoID)
	if err != nil {
		fmt.Println(err)
		return "",  err
	}

	formats := video.Formats.WithAudioChannels()

	if len(formats) == 0 {
		return "", fmt.Errorf("No audio found")
	}

	audioindex := 0

	for i, f := range formats {
		fmt.Println(f.ApproxDurationMs, f.AudioChannels, f.Quality, f.Width, f.ApproxDurationMs, f.LanguageDisplayName(), i)
		if strings.Contains(f.LanguageDisplayName(), "original") {
			audioindex = i
			break
		}
		
	}
		
	stream, _, err := client.GetStream(video, &formats[audioindex])
	if err != nil {
		fmt.Println(err)
		return "",  err
	}
	defer stream.Close()

	fileName := audioPathDirectory + "/" + audioPrefix + videoID +".mp4"
	fmt.Print(fileName)

	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		panic(err)
	}




	return fileName,  nil 
}

func mergeVideoAndAudio (videoFileName string, audioFileName string) (string, error) {
	input1 := ffmpeg.Input(videoFileName)
	input2 := ffmpeg.Input(audioFileName).Audio()

	outputFileName := outputPathDirectory + "/"+ strings.ReplaceAll(videoFileName, videoPathDirectory + "/" + videoPrefix, "")

	err := ffmpeg.Output([]*ffmpeg.Stream{input1, input2}, outputFileName).Run()
	

	
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	file, err := os.Stat(outputFileName)
	if err != nil {
		return "", err
	}
	// get the size
	size := file.Size()

	fmt.Println("file size in bytes: ", size)

	if size > TooBig {
		return FileIsTooBig, nil
	}

	return outputFileName, nil
}

func deleteVideAndAudio () {
	RemoveContents(videoPathDirectory+"/")
	RemoveContents(audioPathDirectory+"/")
}


func RemoveContents(dir string) error {
    d, err := os.Open(dir)
    if err != nil {
        return err
    }
    defer d.Close()
    names, err := d.Readdirnames(-1)
    if err != nil {
        return err
    }
    for _, name := range names {
        err = os.RemoveAll(filepath.Join(dir, name))
        if err != nil {
            return err
        }
    }
    return nil
}
