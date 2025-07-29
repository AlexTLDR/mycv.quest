#let meta = toml("../info.toml")

#import "@preview/grotesk-cv:1.0.5": skill-entry
#import meta.import.fontawesome: *

#let icon = meta.section.icon.skills
#let language = meta.personal.language
#let include-icon = meta.personal.include_icons
#let accent-color = meta.layout.accent_color
#let multicol = true
#let alignment = center

= #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Skills] else if language == "es" [Habilidades]

#v(0pt)

// Check if skills data exists in the TOML
#if "skills" in meta {
  #for skill_group in meta.skills [
    #if type(skill_group) == "dictionary" and "category" in skill_group and "skills" in skill_group [
      === #skill_group.category

      #skill-entry(accent-color, multicol, alignment, skills: (
        ..skill_group.skills.map(skill => [#skill])
      ))

      #v(5pt)
    ] else if type(skill_group) == "string" [
      // If it's a flat array of skills, group them
      #if meta.skills.len() > 0 [
        === Skills

        #skill-entry(accent-color, multicol, alignment, skills: (
          ..meta.skills.map(skill => [#skill])
        ))
      ]
      #break
    ]
  ]
} else if "technical_expertise" in meta [
  === Technical Expertise

  #for expertise in meta.technical_expertise [
    #let skill_name = if "name" in expertise { expertise.name } else { "Skill" }
    #let skill_level = if "level" in expertise { expertise.level } else { 3 }

    // Display skill with level indicators
    #grid(
      columns: (1fr, auto),
      [#skill_name],
      [
        #for i in range(skill_level) [
          #text(fill: rgb(accent-color))[●]
        ]
        #for i in range(5 - skill_level) [
          #text(fill: rgb("#ecf0f1"))[○]
        ]
      ]
    )
  ]
} else if "programming_languages" in meta [
  === Programming Languages

  #skill-entry(accent-color, multicol, alignment, skills: (
    ..meta.programming_languages.map(lang => [#lang])
  ))

  #v(5pt)

  #if "frameworks" in meta [
    === Frameworks & Technologies

    #skill-entry(accent-color, multicol, alignment, skills: (
      ..meta.frameworks.map(fw => [#fw])
    ))

    #v(5pt)
  ]

  #if "tools" in meta [
    === Tools & Platforms

    #skill-entry(accent-color, multicol, alignment, skills: (
      ..meta.tools.map(tool => [#tool])
    ))
  ]
} else [
  // Fallback content based on language
  #if language == "en" [
    === Programming Languages

    #skill-entry(accent-color, multicol, alignment, skills: (
      [Python],
      [JavaScript],
      [Go],
      [TypeScript],
    ))

    #v(5pt)

    === Frameworks & Technologies

    #skill-entry(accent-color, multicol, alignment, skills: (
      [React],
      [Node.js],
      [Docker],
      [Kubernetes],
    ))

    #v(5pt)

    === Tools & Platforms

    #skill-entry(accent-color, multicol, alignment, skills: (
      [Git],
      [AWS],
      [Linux],
      [PostgreSQL],
    ))
  ] else if language == "es" [
    === Lenguajes de Programación

    #skill-entry(accent-color, multicol, alignment, skills: (
      [Python],
      [JavaScript],
      [Go],
      [TypeScript],
    ))

    #v(5pt)

    === Frameworks y Tecnologías

    #skill-entry(accent-color, multicol, alignment, skills: (
      [React],
      [Node.js],
      [Docker],
      [Kubernetes],
    ))

    #v(5pt)

    === Herramientas y Plataformas

    #skill-entry(accent-color, multicol, alignment, skills: (
      [Git],
      [AWS],
      [Linux],
      [PostgreSQL],
    ))
  ]
]
