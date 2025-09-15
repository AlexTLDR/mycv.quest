package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"
)

type ContactInfo struct {
	Name     string
	Title    string
	Email    string
	Website  string
	GitHub   string
	LinkedIn string
	Location string
}

type Experience struct {
	Position  string
	Company   string
	StartDate string
	EndDate   string
	Location  string
	IsRemote  bool
	Bullets   []string
}

type Education struct {
	Degree      string
	Institution string
	StartYear   string
	EndYear     string
	Location    string
}

type TechnicalSkill struct {
	Category string
	Skills   []string
}

type Achievement struct {
	Title       string
	Description string
}

type CVData struct {
	Contact         ContactInfo
	Summary         string
	Experiences     []Experience
	Education       []Education
	TechnicalSkills []TechnicalSkill
	Achievements    []Achievement
	Objective       string
}

func main() {
	cvData := getSampleCVData()

	if err := generateCV(cvData); err != nil {
		log.Fatal(err)
	}

	fmt.Println("CV generated successfully! Check output.pdf")
}

func getSampleCVData() CVData {
	return CVData{
		Contact: ContactInfo{
			Name:     "John Doe",
			Title:    "Software Engineer",
			Email:    "johndoe@example.com",
			Website:  "www.johndoe.com",
			GitHub:   "@johndoe",
			LinkedIn: "johndoe",
			Location: "City, Country",
		},
		Summary:   "Software engineer with 7+ years of experience and a strong foundation in computer science, skilled in developing software for innovative industries. Proficient in JavaScript/TypeScript, Python, and C/C++, with a solid understanding of system architecture and design principles.",
		Objective: "Seeking to advance my skills and build a strong career with a company that values innovation and creativity.",
		Experiences: []Experience{
			{
				Position:  "Lead Software Developer",
				Company:   "Quantum Innovations - QuantumLeap",
				StartDate: "2023 Mar",
				EndDate:   "2024 Jul",
				Location:  "",
				IsRemote:  true,
				Bullets: []string{
					"Spearheaded the development of a cutting-edge quantum computing simulator, optimizing algorithms for performance.",
					"Collaborated with a team to create intuitive user interfaces that simplified complex scientific data for end users.",
				},
			},
			{
				Position:  "Backend Developer",
				Company:   "CloudSync Solutions - SyncManager",
				StartDate: "2024 Aug",
				EndDate:   "present",
				Location:  "",
				IsRemote:  true,
				Bullets: []string{
					"Built scalable backend services for SyncManager, ensuring high availability and performance for cloud synchronization.",
					"Designed and implemented RESTful APIs to facilitate data exchange between clients and servers.",
				},
			},
			{
				Position:  "DevOps Engineer",
				Company:   "AutoTech Dynamics - AutoPilot",
				StartDate: "2022 Feb",
				EndDate:   "2023 Dec",
				Location:  "Denver, USA",
				IsRemote:  false,
				Bullets: []string{
					"Streamlined CI/CD pipelines for the AutoPilot system, enhancing deployment frequency and reliability.",
					"Monitored system performance and implemented improvements for optimized infrastructure.",
				},
			},
			{
				Position:  "Game Developer",
				Company:   "PixelForge Studios - Realm of Adventure",
				StartDate: "2021 Jan",
				EndDate:   "2022 Dec",
				Location:  "Los Angeles, USA",
				IsRemote:  false,
				Bullets: []string{
					"Developed engaging gameplay mechanics and interactive environments using Unity and C\\#.",
					"Collaborated with artists to ensure visual consistency and high-quality game experiences.",
				},
			},
			{
				Position:  "Data Engineer",
				Company:   "Insight Analytics - DataVision",
				StartDate: "2020 Mar",
				EndDate:   "2021 Feb",
				Location:  "Chicago, USA",
				IsRemote:  false,
				Bullets: []string{
					"Engineered data pipelines to aggregate and process large datasets for analytics using Python and Apache Spark.",
					"Developed interactive dashboards for real-time data visualization and business intelligence.",
				},
			},
			{
				Position:  "Quality Assurance Intern",
				Company:   "CodeFix Labs - TestSuite Pro",
				StartDate: "2019 Jan",
				EndDate:   "2019 Dec",
				Location:  "Austin, USA",
				IsRemote:  false,
				Bullets: []string{
					"Assisted in testing software applications for functionality and usability, reporting bugs and feedback.",
					"Gained experience in automated testing frameworks to improve product quality.",
				},
			},
		},
		Education: []Education{
			{
				Degree:      "B.Sc. in Computer Science",
				Institution: "Example University",
				StartYear:   "2015",
				EndYear:     "2019",
				Location:    "City, Country",
			},
			{
				Degree:      "Diploma in IT Specialist",
				Institution: "Technical College",
				StartYear:   "2012",
				EndYear:     "2015",
				Location:    "City, Country",
			},
		},
		TechnicalSkills: []TechnicalSkill{
			{
				Category: "Languages",
				Skills:   []string{"Python", "Java", "React", "Node.js", "Express", "MongoDB", "AWS", "Material UI", "Tailwind CSS"},
			},
			{
				Category: "Methodology/Approach",
				Skills:   []string{"Lean", "Kanban", "Design Thinking", "Test Driven Development", "Pair Programming"},
			},
			{
				Category: "Tools",
				Skills:   []string{"GitHub", "IntelliJ IDEA", "Asana", "Slack", "Adobe XD", "Postman"},
			},
		},
		Achievements: []Achievement{
			{
				Title:       "Best Project Award",
				Description: "Developed an innovative solution for community service management and received recognition from the university.",
			},
		},
	}
}

func generateCV(data CVData) error {
	typstTemplate := `#import "@preview/vantage-cv:1.0.0": vantage-cv

#show: vantage-cv.with(
  name: "{{.Contact.Name}}",
  position: "{{.Contact.Title}}",
  links: (
    (name: "email", link: "{{.Contact.Email}}", icon: "mail"),
    (name: "website", link: "{{.Contact.Website}}", icon: "globe"),
    (name: "github", link: "{{.Contact.GitHub}}", icon: "github"),
    (name: "linkedin", link: "{{.Contact.LinkedIn}}", icon: "linkedin"),
    (name: "location", link: "{{.Contact.Location}}", icon: "map-pin"),
  ),
  tagline: "{{.Summary}}",
  [
    = Objective
    {{.Objective}}

    = Education
    {{range .Education}}
    == {{.Degree}}
    *{{.Institution}}* \
    {{.StartYear}} -- {{.EndYear}} | {{.Location}}

    {{end}}

    = Technical Expertise
    {{range .TechnicalSkills}}
    == {{.Category}}
    {{range .Skills}}{{.}} • {{end}}

    {{end}}

    = Achievements/Certifications
    {{range .Achievements}}
    == {{.Title}}
    {{.Description}}

    {{end}}
  ]
)

= Experience

{{range .Experiences}}
== {{.Position}}
*{{.Company}}* \
{{.StartDate}} -- {{.EndDate}} {{if .Location}}| {{.Location}}{{end}}{{if .IsRemote}} | Remote{{end}}

{{range .Bullets}}
- {{.}}
{{end}}

{{end}}
`

	tmpl, err := template.New("cv").Parse(typstTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	typstFile := "cv.typ"
	file, err := os.Create(typstFile)
	if err != nil {
		return fmt.Errorf("error creating typst file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	if err := compileToPDF(typstFile); err != nil {
		return fmt.Errorf("error compiling to PDF: %v", err)
	}

	return nil
}

func compileToPDF(typstFile string) error {
	outputFile := "output.pdf"

	cmd := exec.Command("typst", "compile", typstFile, outputFile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("typst compilation failed: %v\nOutput: %s", err, output)
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return fmt.Errorf("PDF file was not created")
	}

	fmt.Printf("Successfully compiled %s to %s\n", typstFile, outputFile)
	return nil
}
