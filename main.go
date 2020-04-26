package main

import (
	"log"
	// "math"
  "fmt"
  "time"
  "github.com/spf13/viper"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// Reads the config file
func readConfig() {
  viper.SetConfigName("config") // name of config file (without extension)
  viper.SetConfigType("yaml") // REQUIRED if the config file does not have the extension in the name
  viper.AddConfigPath(".")               // optionally look for config in the working directory
  // viper.AddConfigPath("$HOME/.config/canvas-tui/")              
  err := viper.ReadInConfig() // Find and read the config file
  if err != nil { // Handle errors reading the config file
    panic(fmt.Errorf("Fatal error config file: %s \n", err))
  }
}

// called to generate navigation tabs for courses
func createMainTabPane(courses *[]Course) *widgets.TabPane {
  var titles []string
  titles = append(titles, "Dashboard")
  for _, crs := range *courses {
    if crs.EndAt.IsZero() {
      titles = append(titles, crs.CourseCode)
    }
  }
  tabpane := widgets.NewTabPane(titles...)
	tabpane.Border = true
  return tabpane
}


func  chooseTab(courseMasterGrids []ui.Grid, tabpane *widgets.TabPane, masterGrid *ui.Grid, contentGrid *ui.Grid) {
  // Substitute the current grid for what the user has selected

  // If we click the dashboard 
  if tabpane.ActiveTabIndex == 0 {
    masterGrid.Items[1].Entry = contentGrid
  } else { // for other course pages
    contentGrid = &courseMasterGrids[tabpane.ActiveTabIndex-1]
    masterGrid.Items[1].Entry = contentGrid
  }
  
  ui.Render(masterGrid)
  // log.Panic(contentGrid.Title)
}

func  handleSpace(courseMasterGrids []ui.Grid, courseOverviewGrids []ui.Grid, courseGradeGrids []ui.Grid, courseAnnouncementGrids []ui.Grid, courseSyllabusGrids []ui.Grid, courseAssignmentGrids []ui.Grid, tabpane *widgets.TabPane, masterGrid *ui.Grid, contentGrid *ui.Grid) {
  // Substitute the current grid for what the user has selected

  // If we click the dashboard 
  if tabpane.ActiveTabIndex == 0 {
    masterGrid.Items[1].Entry = contentGrid

  } else { // for other course pages
    // contentGrid = &courseGradeGrids[tabpane.ActiveTabIndex-1]
    // get the currently selected item
    contentGrid = &courseMasterGrids[tabpane.ActiveTabIndex-1]
    item := contentGrid.Items[0].Entry.(*widgets.List).SelectedRow
    itemStr := contentGrid.Items[0].Entry.(*widgets.List).Rows[item]
  
    if itemStr == "Home" {
      contentGrid = &courseMasterGrids[tabpane.ActiveTabIndex-1]
      contentGrid.Items[1].Entry = &courseOverviewGrids[tabpane.ActiveTabIndex-1] 
      masterGrid.Items[1].Entry = contentGrid
    } else if itemStr == "Grades" {
      contentGrid = &courseMasterGrids[tabpane.ActiveTabIndex-1]
      contentGrid.Items[1].Entry = &courseGradeGrids[tabpane.ActiveTabIndex-1]
      masterGrid.Items[1].Entry = contentGrid
    } else if itemStr == "Announcements" {
      contentGrid = &courseMasterGrids[tabpane.ActiveTabIndex-1]
      contentGrid.Items[1].Entry = &courseAnnouncementGrids[tabpane.ActiveTabIndex-1]
      masterGrid.Items[1].Entry = contentGrid
    } else if itemStr == "Syllabus" {
      contentGrid = &courseMasterGrids[tabpane.ActiveTabIndex-1]
      contentGrid.Items[1].Entry = &courseSyllabusGrids[tabpane.ActiveTabIndex-1]
      masterGrid.Items[1].Entry = contentGrid
    } else if itemStr == "Assignments" {
      contentGrid = &courseMasterGrids[tabpane.ActiveTabIndex-1]
      contentGrid.Items[1].Entry = &courseAssignmentGrids[tabpane.ActiveTabIndex-1]
      masterGrid.Items[1].Entry = contentGrid
    } else {
      contentGrid.Items[1].Entry = placeholder()
      masterGrid.Items[1].Entry = contentGrid
    } 



  }
  
  ui.Render(masterGrid)
  // log.Panic(contentGrid.Title)
}


func  menuScroll(coursePages []ui.Grid, tabpane *widgets.TabPane, masterGrid *ui.Grid, contentGrid *ui.Grid, direction string) {
  // Substitute the current grid for what the user has selected
  // Don't try to scroll on the dashboard
  if tabpane.ActiveTabIndex != 0 {
    contentGrid = &coursePages[tabpane.ActiveTabIndex-1]
    l := contentGrid.Items[0].Entry.(*widgets.List)
    if direction == "down" {
      l.ScrollDown()
    } else if direction == "up"{
      l.ScrollUp()
    }
  }
  ui.Render(masterGrid)
}

// called if master grid needs to be updated
func updateMasterGrid(masterGrid *ui.Grid, tabpane *widgets.TabPane, contentGrid *ui.Grid) *ui.Grid {
  ui.Clear()
  // defining master grid layout
  masterGrid.Set(
    ui.NewRow(1.0/20,
      ui.NewCol(1.0, tabpane),
    ),
    ui.NewRow(19.0/20,
      ui.NewCol(1.0/1, contentGrid),
    ),
  )
  // ui.Render(masterGrid)
  return masterGrid
}

func main() {
  
  // Initialize temui
  if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

  // get the main config
  readConfig()

  var courses *[]Course = fetchCourses()
  
  // declare master grid and set terminal dimensions
	masterGrid := ui.NewGrid()
  masterGrid.Title = "Master Grid"
	termWidth, termHeight := ui.TerminalDimensions()
	masterGrid.SetRect(0, 0, termWidth, termHeight)

  // declare tab widget
  tabpane := createMainTabPane(courses)

  // contentGrid := createDashboardGrid("front page")
  dashboard := createDashboardGrid(courses)
  contentGrid := dashboard

  // Do the initial drawing of the main dash
  masterGrid = updateMasterGrid(masterGrid, tabpane, contentGrid)

  // one list of assignments per course
  var assignmentsMatrix [][]Assignment

  // one list of announcements per course
  var announcementMatrix [][]Announcement

  // one master grid per course
  var courseMasterGrids []ui.Grid

  // one grade grid per course 
  var courseGradeGrids []ui.Grid

  // one announcement grid per course 
  var courseAnnouncementGrids []ui.Grid

  // one course overview grid per course 
  var courseOverviewGrids []ui.Grid

  // one syllabus grid per course 
  var courseSyllabusGrids []ui.Grid

  // one assignment grid per course
  var courseAssignmentGrids []ui.Grid


  // first fetch all the assignments to reduce redundant API calls
  for _, crs := range *courses {
    if crs.EndAt.IsZero() {
      assignmentsMatrix = append(assignmentsMatrix, *fetchAssignments(crs.ID))
      announcementMatrix = append(announcementMatrix, *fetchAnnouncements(crs.ID))
    }
  }
  i := 0
  for _, crs := range *courses {
    if crs.EndAt.IsZero() {
      courseMasterGrids = append(courseMasterGrids, *createCourseGrid(crs, &assignmentsMatrix[i], &announcementMatrix[i]))
      courseOverviewGrids = append(courseOverviewGrids, *createCourseOverviewGrid(crs, &assignmentsMatrix[i], &announcementMatrix[i]))
      courseGradeGrids = append(courseGradeGrids, *createGradeGrid(crs))
      courseAnnouncementGrids = append(courseAnnouncementGrids, *createAnnouncementGrid(crs, &announcementMatrix[i]))
      courseSyllabusGrids = append(courseSyllabusGrids, *createSyllabusGrid(crs))
      courseAssignmentGrids = append(courseAssignmentGrids, *createAssignmentGrid(crs, &assignmentsMatrix[i]))
      i++
    }
  }


  // Event polling loop
  tickerCount := 1
	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C
	for {
    select {
    case e := <-uiEvents:
      switch e.ID {
      case "q", "<C-c>":
        return
      case "h":
        tabpane.FocusLeft() // changes the currently selected tab
        ui.Render(tabpane) // quickly redraws tabpane
      case "l":
        tabpane.FocusRight()
        ui.Render(tabpane)
      case "j":
        menuScroll(courseMasterGrids, tabpane, masterGrid, contentGrid, "down")
      case "k":
        menuScroll(courseMasterGrids, tabpane, masterGrid, contentGrid, "up")
      case "<Enter>":
        chooseTab(courseMasterGrids, tabpane, masterGrid, contentGrid)
      case "<Space>":
        handleSpace(courseMasterGrids, courseOverviewGrids, courseGradeGrids, courseAnnouncementGrids, courseSyllabusGrids, courseAssignmentGrids, tabpane, masterGrid, contentGrid)
      case "<Resize>":
				payload := e.Payload.(ui.Resize)
				masterGrid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(masterGrid)
      }
		case <-ticker:
      ui.Render(masterGrid)
			tickerCount++
		}
	}

}
