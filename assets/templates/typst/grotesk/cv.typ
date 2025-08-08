#let meta = toml("./info.toml")

// Import the grotesk-cv template
#import "@preview/grotesk-cv:1.0.5": cv, experience-entry, education-entry, skill-entry, language-entry, reference-entry
#import meta.import.fontawesome: *

// Load image
#let photo = image(width: 100%, height: 100%, "./img/" + meta.personal.profile_image)

// Helper function to safely get values
#let get-value(obj, key, default: "") = {
  if key in obj {
    obj.at(key)
  } else {
    default
  }
}

// Generate profile/summary section
#let profile-section = {
  let icon = meta.section.icon.profile
  let language = meta.personal.language
  let include-icon = meta.personal.include_icons

  [
    = #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Summary] else if language == "es" [Resumen]

    #v(5pt)

    #{
      if "summary" in meta and meta.summary != none and meta.summary != "" {
        meta.summary
      } else {
        // Fallback content based on language
        if language == "en" {
          [Experienced Software Engineer with expertise in modern technologies and best practices. Passionate about creating efficient solutions and contributing to innovative projects in dynamic environments.]
        } else if language == "es" {
          [Ingeniero de Software experimentado con experiencia en tecnologías modernas y mejores prácticas. Apasionado por crear soluciones eficientes y contribuir a proyectos innovadores en entornos dinámicos.]
        } else {
          [Experienced Software Engineer with expertise in modern technologies and best practices. Passionate about creating efficient solutions and contributing to innovative projects in dynamic environments.]
        }
      }
    }
  ]
}

// Generate experience section
#let experience-section = {
  let icon = meta.section.icon.experience
  let language = meta.personal.language
  let include-icon = meta.personal.include_icons

  [
    = #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Experience] else if language == "es" [Experiencia]

    #v(5pt)

    #{
      if "experience" in meta {
        for exp in meta.experience {
          let title = get-value(exp, "position", default: get-value(exp, "title", default: "Position"))
          let company = if "company" in exp {
            if type(exp.company) == dictionary and "name" in exp.company {
              exp.company.name
            } else {
              str(exp.company)
            }
          } else { "Company" }
          let location = get-value(exp, "location", default: "")
          let date_from = get-value(exp, "from", default: "")
          let date_to = get-value(exp, "to", default: "")
          let date_range = if date_from != "" and date_to != "" { date_from + " - " + date_to } else { "" }
          let description = get-value(exp, "description", default: ())

          experience-entry(
            title: title,
            company: company,
            location: location,
            date: date_range
          )

          if type(description) == array and description.len() > 0 {
            for item in description {
              [• #item]
              linebreak()
            }
          }

          v(0.5em)
        }
      }
    }
  ]
}

// Generate education section
#let education-section = {
  let icon = meta.section.icon.education
  let language = meta.personal.language
  let include-icon = meta.personal.include_icons

  [
    = #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Education] else if language == "es" [Educación]

    #v(5pt)

    #{
      if "education" in meta {
        for edu in meta.education {
          let degree = get-value(edu, "degree", default: "Degree")
          let institution = if "institution" in edu {
            edu.institution
          } else if "place" in edu {
            if type(edu.place) == dictionary and "name" in edu.place {
              edu.place.name
            } else {
              str(edu.place)
            }
          } else { "Institution" }
          let location = get-value(edu, "location", default: "")
          let date_from = get-value(edu, "from", default: "")
          let date_to = get-value(edu, "to", default: "")
          let date_range = if date_from != "" and date_to != "" { date_from + " - " + date_to } else { "" }
          let major = get-value(edu, "major", default: "")

          education-entry(
            institution: institution,
            degree: degree + if major != "" { " in " + major } else { "" },
            location: location,
            date: date_range
          )

          v(0.3em)
        }
      }
    }
  ]
}

// Generate skills section
#let skills-section = {
  let icon = meta.section.icon.skills
  let language = meta.personal.language
  let include-icon = meta.personal.include_icons
  let accent-color = meta.layout.accent_color

  [
    = #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Skills] else if language == "es" [Habilidades]

    #v(5pt)

    #{
      if "skills" in meta {
        if type(meta.skills) == array {
          // Simple array of skills
          for skill in meta.skills {
            skill-entry(accent-color, true, center, skills: (skill,))
          }
        } else {
          // Complex skills structure
          for skill_group in meta.skills {
            if type(skill_group) == dictionary and "category" in skill_group and "skills" in skill_group {
              [=== #skill_group.category]
              skill-entry(accent-color, true, center, skills: skill_group.skills)
            }
          }
        }
      }
    }
  ]
}

// Generate languages section
#let languages-section = {
  let icon = meta.section.icon.languages
  let language = meta.personal.language
  let include-icon = meta.personal.include_icons

  [
    = #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Languages] else if language == "es" [Idiomas]

    #v(5pt)

    #{
      if "languages" in meta {
        for lang in meta.languages {
          let lang_name = get-value(lang, "name", default: get-value(lang, "language", default: "Language"))
          let proficiency = get-value(lang, "proficiency", default: get-value(lang, "level", default: ""))
          language-entry(language: lang_name, proficiency: proficiency)
        }
      } else {
        // Default languages
        if language == "en" {
          language-entry(language: "English", proficiency: "Native")
          language-entry(language: "Spanish", proficiency: "Fluent")
        } else if language == "es" {
          language-entry(language: "Inglés", proficiency: "Nativo")
          language-entry(language: "Español", proficiency: "Fluido")
        }
      }
    }
  ]
}

// Generate other experience section
#let other-experience-section = {
  let icon = meta.section.icon.other_experience
  let language = meta.personal.language
  let include-icon = meta.personal.include_icons

  [
    = #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Other] else if language == "es" [Otra experiencia]

    #v(5pt)

    #{
      if "other_experience" in meta {
        for exp in meta.other_experience {
          let title = get-value(exp, "title", default: "Experience")
          let company = get-value(exp, "company", default: "")
          let location = get-value(exp, "location", default: "")
          let date = get-value(exp, "date", default: "")

          experience-entry(title: title, company: company, location: location, date: date)
        }
      }
    }
  ]
}

// Generate references section
#let references-section = {
  let icon = meta.section.icon.references
  let language = meta.personal.language
  let include-icon = meta.personal.include_icons

  [
    = #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [References] else if language == "es" [Referencias]

    #v(5pt)

    #{
      if "references" in meta {
        for ref in meta.references {
          let name = get-value(ref, "name", default: "Reference")
          let company = get-value(ref, "company", default: "")
          let telephone = get-value(ref, "telephone", default: get-value(ref, "phone", default: ""))
          let email = get-value(ref, "email", default: "")

          reference-entry(
            name: name,
            company: company,
            telephone: telephone,
            email: email
          )
        }
      }
    }
  ]
}

// Generate personal section
#let personal-section = {
  let icon = meta.section.icon.personal
  let language = meta.personal.language
  let include-icon = meta.personal.include_icons

  [
    = #if include-icon [#fa-icon(icon) #h(5pt)] #if language == "en" [Personality] else if language == "es" [Personalidad]

    #v(5pt)

    #{
      if "personal" in meta and "traits" in meta.personal {
        for trait in meta.personal.traits {
          [• #trait]
          linebreak()
        }
      } else {
        // Default personality traits
        if language == "en" {
          [• Analytic thinking]
          linebreak()
          [• Quality conscious]
          linebreak()
          [• Good communicator]
          linebreak()
          [• Independent]
          linebreak()
          [• Team player]
          linebreak()
          [• Preemptive]
          linebreak()
          [• Eager to learn]
        } else if language == "es" {
          [• Pensamiento analítico]
          linebreak()
          [• Consciente de la calidad]
          linebreak()
          [• Buen comunicador]
          linebreak()
          [• Independiente]
          linebreak()
          [• Jugador de equipo]
          linebreak()
          [• Preventivo]
          linebreak()
          [• Ansioso por aprender]
        }
      }
    }
  ]
}

// Define the pane contents
#let left-pane = profile-section + experience-section + education-section
#let right-pane = skills-section + languages-section + other-experience-section + references-section

// Apply the cv template
#show: cv.with(
  meta,
  photo: photo,
  use-photo: true,
  left-pane: left-pane,
  right-pane: right-pane,
  left-pane-proportion: eval(meta.layout.left_pane_width),
)
