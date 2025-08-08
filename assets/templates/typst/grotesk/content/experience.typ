#let meta = toml("../info.toml")

#import "@preview/grotesk-cv:1.0.5": experience-entry
#import meta.import.fontawesome: *

#let icon = meta.section.icon.experience
#let language = meta.personal.language
#let include-icon = meta.personal.include_icons

= #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Experience] else if language == "es" [Experiencia]

#v(5pt)

// Check if experience data exists in the TOML
#if "experience" in meta {
  #for exp in meta.experience [
    #let title = if "position" in exp { exp.position } else if "title" in exp { exp.title } else { "Position" }
    #let company = if "company" in exp {
      if type(exp.company) == "dictionary" and "name" in exp.company {
        exp.company.name
      } else {
        exp.company
      }
    } else { "Company" }
    #let location = if "location" in exp { exp.location } else { "" }
    #let from_date = if "from" in exp { exp.from } else { "" }
    #let to_date = if "to" in exp { exp.to } else { "" }
    #let date_range = if from_date != "" and to_date != "" { from_date + " - " + to_date } else if from_date != "" { from_date } else { "Date" }

    #experience-entry(
      title: [#title],
      date: [#date_range],
      company: [#company],
      location: [#location],
    )

    // Add description points if they exist
    #if "description" in exp and type(exp.description) == "array" [
      #for point in exp.description [
        - #point
      ]
    ] else if "description" in exp [
      - #exp.description
    ]

    // Add technologies/skills if they exist
    #if "technologies" in exp and type(exp.technologies) == "array" [
      #v(2pt)
      #text(size: 9pt, style: "italic")[Technologies: #exp.technologies.join(", ")]
    ] else if "tags" in exp and type(exp.tags) == "array" [
      #v(2pt)
      #text(size: 9pt, style: "italic")[Technologies: #exp.tags.join(", ")]
    ]

    #v(5pt)
  ]
} else [
  // Fallback if no experience data
  #text(style: "italic")[No experience data provided]
]
