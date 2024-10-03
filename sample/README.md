
`$ ffmpeg -i MooveKa_-_Rockabilly_Punk_Rock.mp3 -map 0:a:0 -ar 44100 -ac 2 -c:a pcm_f32le -f f32le test_f32.raw`

`$ ffmpeg -i MooveKa_-_Rockabilly_Punk_Rock.mp3 -map 0:a:0 -ar 44100 -ac 2 -c:a pcm_s16le -f s16le test_s16.raw`
