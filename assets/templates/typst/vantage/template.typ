#import "vantage-typst.typ": vantage, term, skill, styled-link

// Main function for vantage CV template
#let vantage-cv(data) = {
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

  // Extract contacts information
  let contacts = data.at("contacts", default: (:))
  let name = get-value(data, ("contacts", "name"), default: "")
  let position = data.at("position", default: "")
  let tagline = data.at("tagline", default: "")

  // Build links array
  let links = ()

  // Email link
  if has-value(data, ("contacts", "email")) {
    let email = get-value(data, ("contacts", "email"))
    links.push((name: "email", link: "mailto:" + email))
  }

  // Website link
  if has-value(data, ("contacts", "website", "url")) {
    let website = get-value(data, ("contacts", "website"))
    let url = website.at("url", default: "")
    let display = website.at("displayText", default: url)
    links.push((name: "website", link: url, display: display))
  }

  // GitHub link
  if has-value(data, ("contacts", "github", "url")) {
    let github = get-value(data, ("contacts", "github"))
    let url = github.at("url", default: "")
    let display = github.at("displayText", default: url)
    links.push((name: "github", link: url, display: display))
  }

  // LinkedIn link
  if has-value(data, ("contacts", "linkedin", "url")) {
    let linkedin = get-value(data, ("contacts", "linkedin"))
    let url = linkedin.at("url", default: "")
    let display = linkedin.at("displayText", default: url)
    links.push((name: "linkedin", link: url, display: display))
  }

  // Location
  if has-value(data, ("contacts", "address")) {
    let address = get-value(data, ("contacts", "address"))
    links.push((name: "location", link: "", display: address))
  } else if has-value(data, ("contacts", "location")) {
    let location = get-value(data, ("contacts", "location"))
    links.push((name: "location", link: "", display: location))
  }

  // Call the vantage template with our data
  vantage(
    name: name,
    position: position,
    links: links,
    tagline: tagline,

    // Left column content
    [
      #if "jobs" in data and data.jobs.len() > 0 [
        == Experience

        #for job in data.jobs [
          === #job.at("position", default: "") \

          #let company = job.at("company", default: (:))
          #let company-name = company.at("name", default: "")
          #let company-link = company.at("link", default: "")

          _#if company-link != "" [
            #link(company-link)[#company-name]
          ] else [
            #company-name
          ]_

          #if "product" in job [
            #let product = job.product
            #let product-name = product.at("name", default: "")
            #let product-link = product.at("link", default: "")
            - #if product-link != "" [
              #styled-link(product-link, product-name)
            ] else [
              #product-name
            ]
          ] \

          #let job-from = job.at("from", default: "")
          #let job-to = job.at("to", default: "")
          #let job-location = job.at("location", default: "")
          #term[#job-from --- #job-to][#job-location]

          #if "description" in job and job.description.len() > 0 [
            #for point in job.description [
              - #point
            ]
          ]

          #v(0.5em)
        ]
      ]
    ],

    // Right column content
    [
      #if "objective" in data and data.objective != "" [
        == Objective
        #data.objective
        #v(0.8em)
      ]

      #if "education" in data and data.education.len() > 0 [
        == Education

        #for edu in data.education [
          #let place = edu.at("place", default: (:))
          #let place-name = place.at("name", default: "")
          #let place-link = place.at("link", default: "")

          === #if place-link != "" [
            #link(place-link)[#place-name]
          ] else [
            #place-name
          ] \

          #let edu-from = edu.at("from", default: "")
          #let edu-to = edu.at("to", default: "")
          #let edu-location = edu.at("location", default: "")
          #edu-from - #edu-to #h(1fr) #edu-location

          #let degree = edu.at("degree", default: "")
          #let major = edu.at("major", default: "")
          #degree in #major

          #v(0.3em)
        ]
        #v(0.5em)
      ]

      #if "technical_expertise" in data and data.technical_expertise.len() > 0 [
        == Technical Expertise

        #for expertise in data.technical_expertise [
          #let skill-name = expertise.at("name", default: "")
          #let skill-level = expertise.at("level", default: 1)
          #skill(skill-name, skill-level)
        ]
        #v(0.5em)
      ]

      #if "skills" in data and data.skills.len() > 0 [
        == Skills/Exposure

        #for skill-item in data.skills [
          • #skill-item
        ]
        #v(0.5em)
      ]

      #if "methodology" in data and data.methodology.len() > 0 [
        == Methodology/Approach
        #for method in data.methodology [
          • #method
        ]
        #v(0.5em)
      ]

      #if "tools" in data and data.tools.len() > 0 [
        == Tools
        #for tool in data.tools [
          • #tool
        ]
        #v(0.5em)
      ]

      #if "achievements" in data and data.achievements.len() > 0 [
        == Achievements/Certifications

        #for achievement in data.achievements [
          #let ach-name = achievement.at("name", default: "")
          #let ach-description = achievement.at("description", default: "")

          === #ach-name
          \
          #ach-description
          #v(0.3em)
        ]
      ]
    ]
  )
}
