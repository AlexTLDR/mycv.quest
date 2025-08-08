#let meta = toml("./info.toml")

#import "@preview/grotesk-cv:1.0.5": cv
#let photo = image(width: 100%, height: 100%, "./img/" + meta.personal.profile_image)

#show: cv.with(
  meta,
  photo: photo,
  use-photo: meta.personal.use_photo,
  left-pane: [],
  right-pane: [],
  left-pane-proportion: eval(meta.layout.left_pane_width),
)
