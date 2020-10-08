# Cloud-gaming 

## Try it!
**http://cloud-gaming.southeastasia.cloudapp.azure.com:8000**

## Intro
**Cloud gaming proof of concept implemented in Golang using WebRTC**
* Playing 2-player NES games on browser using emulator provided by https://github.com/fogleman/nes
* Go WebRTC implementation developed by https://github.com/pion/webrtc
* Other key dependencies:
    * Websocket: https://github.com/gorilla/websocket
    * VP8 video encoding: https://github.com/poi5305/go-yuv2webRTC/vpx-encoder
* Currently working only on Chrome

## How to run
**Local:**
* Install dependency:
    * Ubuntu: apt-get install -y pkg-config libvpx-dev
    * MacOS: brew install libvpx pkg-config 
* Require Golang SDK version 1.15
* Run: go run cmd/worker/*

**Docker compose:**

**Currently not working on Docker on Mac. I'm trying to figuring out.**
* Run: docker-compose up --build