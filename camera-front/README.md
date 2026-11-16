# The Camera Front-end
## How to build the image
```
docker build -t camera-front .
```

## How to run the image
```
docker run -d --rm -p 9504:9504 camera-front:latest 
```