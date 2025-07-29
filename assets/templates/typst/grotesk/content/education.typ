#let meta = toml("../info.toml")

#import "@preview/grotesk-cv:1.0.5": education-entry
#import meta.import.fontawesome: *

#let icon = meta.section.icon.education
#let language = meta.personal.language
#let include-icon = meta.personal.include_icons

= #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Education] else if language == "es" [Educación]

#v(5pt)

// Check if education data exists in the TOML
#if "education" in meta {
  #for edu in meta.education [
    #let degree = if "degree" in edu { edu.degree } else { "Degree" }
    #let institution = if "institution" in edu {
      edu.institution
    } else if "place" in edu {
      if type(edu.place) == "dictionary" and "name" in edu.place {
        edu.place.name
      } else {
        edu.place
      }
    } else { "Institution" }
    #let from_date = if "from" in edu { edu.from } else { "" }
    #let to_date = if "to" in edu { edu.to } else { "" }
    #let date_range = if from_date != "" and to_date != "" { from_date + " - " + to_date } else if from_date != "" { from_date } else { "Date" }

    #education-entry(
      degree: [#degree],
      date: [#date_range],
      institution: [#institution],
    )

    // Add field of study if available
    #if "field" in edu or "major" in edu [
      #let field = if "field" in edu { edu.field } else { edu.major }
      - *Field:* #field
    ]

    // Add GPA if available
    #if "gpa" in edu [
      - *GPA:* #edu.gpa
    ]

    // Add honors if available
    #if "honors" in edu [
      - *Honors:* #edu.honors
    ]

    // Add thesis if available
    #if "thesis" in edu [
      - *Thesis:* _"#edu.thesis"_
    ]

    // Add description if available
    #if "description" in edu and type(edu.description) == "array" [
      #for point in edu.description [
        - #point
      ]
    ] else if "description" in edu [
      - #edu.description
    ]

    // Add research focus if available
    #if "research_focus" in edu [
      - *Research focus:* #edu.research_focus
    ]

    #v(5pt)
  ]
} else [
  // Fallback if no education data
  #text(style: "italic")[No education data provided]
]
