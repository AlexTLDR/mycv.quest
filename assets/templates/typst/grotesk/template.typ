#import "@preview/fontawesome:0.1.0": *

// Grotesk CV template that works with vantage data structure
#let grotesk-cv(data) = {
  // Helper function to safely get nested values
  let get-value(obj, path, default: "") = {
    let current = obj
    for key in path {
      if type(current) == dictionary and key in current {
        current = current.at(key)
      } else {
        return default
      }
    }
    current
  }

  // Helper function to check if value exists and is not empty
  let has-value(obj, path) = {
    let val = get-value(obj, path)
    val != none and val != ""
  }

  // Page setup
  set page(
    paper: "a4",
    margin: (left: 0.7in, right: 0.7in, top: 0.7in, bottom: 0.7in),
  )

  set text(
    font: ("Source Sans Pro", "Liberation Sans", "Arial"),
    size: 10pt,
    fill: rgb("2c3e50"),
  )

  set par(justify: true)

  // Extract contacts information
  let contacts = data.at("contacts", default: (:))
  let name = get-value(data, ("contacts", "name"), default: "")
  let title = get-value(data, ("contacts", "title"), default: "")
  let position = data.at("position", default: title)
  let tagline = data.at("tagline", default: "")

  // Header section
  align(center)[
    #text(size: 24pt, weight: "bold", fill: rgb("34495e"))[
      #upper(name)
    ]

    #if position != "" [
      #v(0.3em)
      #text(size: 14pt, style: "italic", fill: rgb("7f8c8d"))[
        #position
      ]
    ]
  ]

  v(0.5em)

  // Contact information
  let contact-items = ()

  if has-value(data, ("contacts", "email")) {
    let email = get-value(data, ("contacts", "email"))
    contact-items.push([#fa-envelope() #link("mailto:" + email)[#email]])
  }

  if has-value(data, ("contacts", "address")) {
    let address = get-value(data, ("contacts", "address"))
    contact-items.push([#fa-map-marker() #address])
  } else if has-value(data, ("contacts", "location")) {
    let location = get-value(data, ("contacts", "location"))
    contact-items.push([#fa-map-marker() #location])
  }

  if has-value(data, ("contacts", "linkedin", "url")) {
    let linkedin = get-value(data, ("contacts", "linkedin"))
    let url = linkedin.at("url", default: "")
    let display = linkedin.at("displayText", default: "LinkedIn")
    contact-items.push([#fa-linkedin() #link(url)[#display]])
  }

  if has-value(data, ("contacts", "github", "url")) {
    let github = get-value(data, ("contacts", "github"))
    let url = github.at("url", default: "")
    let display = github.at("displayText", default: "GitHub")
    contact-items.push([#fa-github() #link(url)[#display]])
  }

  if has-value(data, ("contacts", "website", "url")) {
    let website = get-value(data, ("contacts", "website"))
    let url = website.at("url", default: "")
    let display = website.at("displayText", default: "Website")
    contact-items.push([#fa-globe() #link(url)[#display]])
  }

  // Display contact items
  if contact-items.len() > 0 {
    align(center)[
      #grid(
        columns: contact-items.len(),
        column-gutter: 1.5em,
        ..contact-items
      )
    ]
  }

  v(1em)
  line(length: 100%, stroke: 1pt + rgb("bdc3c7"))
  v(1em)

  // Two-column layout
  grid(
    columns: (1.2fr, 1fr),
    column-gutter: 1.5em,

    // Left column
    [
      #if tagline != "" [
        == #fa-user() Summary
        #par(justify: true)[#tagline]
        #v(0.8em)
      ]

      #if "jobs" in data and data.jobs.len() > 0 [
        == #fa-briefcase() Experience

        #for job in data.jobs [
          #let job-position = job.at("position", default: "")
          #let company = job.at("company", default: (:))
          #let company-name = company.at("name", default: "")
          #let company-link = company.at("link", default: "")
          #let job-from = job.at("from", default: "")
          #let job-to = job.at("to", default: "")
          #let job-location = job.at("location", default: "")

          === #job-position

          #grid(
            columns: (1fr, auto),
            [
              #if company-link != "" [
                #link(company-link)[*#company-name*]
              ] else [
                *#company-name*
              ]
            ],
            [
              #text(size: 9pt, fill: rgb("7f8c8d"))[
                #job-from -- #job-to
                #if job-location != "" [ | #job-location]
              ]
            ]
          )

          #if "product" in job [
            #let product = job.product
            #let product-name = product.at("name", default: "")
            #let product-link = product.at("link", default: "")
            #text(size: 9pt, style: "italic")[
              Product: #if product-link != "" [
                #link(product-link)[#product-name]
              ] else [
                #product-name
              ]
            ]
          ]

          #if "description" in job and job.description.len() > 0 [
            #for point in job.description [
              • #point
            ]
          ]

          #if "tags" in job and job.tags.len() > 0 [
            #v(0.2em)
            #text(size: 8pt, fill: rgb("3498db"))[
              *Technologies:* #job.tags.join(", ")
            ]
          ]

          #v(0.6em)
        ]
      ]
    ],

    // Right column
    [
      #if "education" in data and data.education.len() > 0 [
        == #fa-graduation-cap() Education

        #for edu in data.education [
          #let place = edu.at("place", default: (:))
          #let place-name = place.at("name", default: "")
          #let place-link = place.at("link", default: "")
          #let degree = edu.at("degree", default: "")
          #let major = edu.at("major", default: "")
          #let track = edu.at("track", default: "")
          #let edu-from = edu.at("from", default: "")
          #let edu-to = edu.at("to", default: "")
          #let edu-location = edu.at("location", default: "")

          === #if place-link != "" [
            #link(place-link)[#place-name]
          ] else [
            #place-name
          ]

          #grid(
            columns: (1fr, auto),
            [
              #degree #if major != "" [ in #major]
              #if track != "" [
                #linebreak()
                #text(size: 9pt, style: "italic")[Track: #track]
              ]
            ],
            [
              #text(size: 9pt, fill: rgb("7f8c8d"))[
                #edu-from -- #edu-to
                #if edu-location != "" [
                  #linebreak()
                  #edu-location
                ]
              ]
            ]
          )

          #v(0.5em)
        ]
        #v(0.5em)
      ]

      #if "technical_expertise" in data and data.technical_expertise.len() > 0 [
        == #fa-cogs() Technical Expertise

        #for expertise in data.technical_expertise [
          #let skill-name = expertise.at("name", default: "")
          #let skill-level = expertise.at("level", default: 1)

          #grid(
            columns: (1fr, auto),
            [#skill-name],
            [
              #for i in range(skill-level) [
                #text(fill: rgb("e74c3c"))[●]
              ]
              #for i in range(5 - skill-level) [
                #text(fill: rgb("ecf0f1"))[●]
              ]
            ]
          )
        ]
        #v(0.5em)
      ]

      #if "skills" in data and data.skills.len() > 0 [
        == #fa-code() Skills

        #let skill-groups = ()
        #let current-group = ()

        #for skill in data.skills [
          #current-group.push(skill)
          #if current-group.len() == 3 [
            #skill-groups.push(current-group)
            #current-group = ()
          ]
        ]

        #if current-group.len() > 0 [
          #skill-groups.push(current-group)
        ]

        #for group in skill-groups [
          #text(size: 9pt)[#group.join(" • ")]
          #linebreak()
        ]

        #v(0.5em)
      ]

      #if "methodology" in data and data.methodology.len() > 0 [
        == #fa-tasks() Methodology
        #for method in data.methodology [
          • #method
        ]
        #v(0.5em)
      ]

      #if "tools" in data and data.tools.len() > 0 [
        == #fa-wrench() Tools
        #text(size: 9pt)[#data.tools.join(" • ")]
        #v(0.5em)
      ]

      #if "achievements" in data and data.achievements.len() > 0 [
        == #fa-trophy() Achievements

        #for achievement in data.achievements [
          #let ach-name = achievement.at("name", default: "")
          #let ach-description = achievement.at("description", default: "")

          === #ach-name
          #text(size: 9pt)[#ach-description]
          #v(0.4em)
        ]
      ]
    ]
  )
}
