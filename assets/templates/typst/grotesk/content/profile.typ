#let meta = toml("../info.toml")
#import meta.import.fontawesome: *

#let icon = meta.section.icon.profile
#let language = meta.personal.language
#let include-icon = meta.personal.include_icons

// = Summary
= #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Summary] else if language == "es" [Resumen]

#v(5pt)

// Check if summary/tagline data exists in the TOML
#if "summary" in meta {
  #meta.summary
} else if "tagline" in meta {
  #meta.tagline
} else if "professional_summary" in meta {
  #meta.professional_summary
} else if "profile" in meta {
  #meta.profile
} else [
  // Fallback content based on language
  #if language == "en" [
    Experienced Software Engineer with expertise in modern technologies and best practices. Passionate about creating efficient solutions and contributing to innovative projects in dynamic environments.
  ] else if language == "es" [
    Ingeniero de Software experimentado con experiencia en tecnologías modernas y mejores prácticas. Apasionado por crear soluciones eficientes y contribuir a proyectos innovadores en entornos dinámicos.
  ] else [
    Experienced Software Engineer with expertise in modern technologies and best practices. Passionate about creating efficient solutions and contributing to innovative projects in dynamic environments.
  ]
]
