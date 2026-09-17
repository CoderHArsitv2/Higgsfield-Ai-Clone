# Showcase footage

Placeholder footage from [Pexels](https://www.pexels.com), free for commercial
use with no attribution required. Credited here anyway.

| File | Source |
| --- | --- |
| `rain-city.mp4` | pexels.com/video/30169690 |
| `neon-street.mp4` | pexels.com/video/29834228 |
| `coastline.mp4` | pexels.com/video/7913483 |
| `desert.mp4` | pexels.com/video/19376556 |
| `portrait.mp4` | pexels.com/video/19863106 |
| `studio.mp4` | pexels.com/video/3129671 |

Each clip is a 5-second silent loop, transcoded to ~200-450KB:

```
ffmpeg -ss 1 -t 5 -i SOURCE.mp4 \
  -vf "scale=1280:-2,crop=1280:720,fps=24" \
  -c:v libx264 -crf 31 -preset slow -pix_fmt yuv420p \
  -movflags +faststart -an out.mp4
```

**To swap in your own footage:** keep the same filenames and the site picks it
up with no code change. Filenames are referenced in `HeroMedia.tsx`,
`Showcase.tsx` and `LoopSection.tsx`.
