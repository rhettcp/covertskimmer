# CovertSkimmer

Looking for a way to get the latest images from your Covert Camera? Here is a handy little lib to help!

## The Problem

Covert Wireless currently doesn't provide an API to access camera data or images. You can use their app or the web portal. So if you had ideas of an alerting app, image processing, or just an image dashboard... there wasn't a way to do so.

## The Solution

After digging through the online api calls, it seems as if the page is generated server side so there are no api endpoints to hit.

Since this was the case, a good old fashioned web scrape seemed okay. This lib scrapes the page for the image urls (since they are in S3).

## Usage

An example speaks volumes...
```
client, err := covertskimmer.NewCovertClient(username, password)
if err != nil {
	panic(err)
}
for _, c := range client.GetCameras() {
    fmt.Println("Images: ", client.GetImageList(*c))
}
```