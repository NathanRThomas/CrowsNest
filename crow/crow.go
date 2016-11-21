/*! \file crow.go
  \brief Handles mostly initialization and setup stuff from our config file for the service
*/

package crow

import (
    "fmt"
    "os"
    "encoding/json"
    "net/url"
    "regexp"
    )

type capture_t struct {
    Alias           string  `json:"alias"`
    Regex           string  `json:"regex"`
    Exists          bool    `json:"exists"`
    Missing         bool    `json:"missing"`
    Max             int     `json:"max"`
    Min             int     `json:"min"`
}

type egg_t struct {
    Alias           string  `json:"alias"`
    Url             string  `json:"url"`
    Interval        int     `json:"interval"`
    Class           int     `json:"class"`
    Errors          []capture_t  `json:"errors"`
    Warnings        []capture_t  `json:"warnings"`
}

type crew_t struct {
    Alias           string  `json:"alias"`
    Phone           string  `json:"phone"`
    Email           string  `json:"email"`
    ClassMask       int     `json:"class_mask"`
}

type Crow_c struct {
    eggs                []egg_t
    crew                []crew_t
    elapsedIntervals    int
    errEggs             map[int]bool
    warnEggs            map[int]bool
    
    crowUrl             crow_url_c       //this checks the urls for us for issues
    crowSquawk          crow_squawk_c   //this handles notifing the crew of an issue
}

//-------------------------------------------------------------------------------------------------------------------------//
//----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Validates that a url is of a valid format
 */
func (c Crow_c) validateUrl (inUrl string) {
    _, err := url.Parse(inUrl)
    if err != nil { panic(err) }    //this is a cause for concern
}

/*! \brief Validates the syntax of the regex for capturing the page output
 */
func (c Crow_c) validateRegex (in string) {
    if len(in) > 0 {
        regexp.MustCompile(in)
    }
}

//-------------------------------------------------------------------------------------------------------------------------//
//----- PUBLIC FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Main entry point.  This will read our config files and make sure we can start running
 */
func (c *Crow_c) Init () (error) {
    
    //Read in the eggs
    configFile, err := os.Open("eggs.json") //try the file
    
	if err == nil {
        defer configFile.Close()
		jsonParser := json.NewDecoder(configFile)
		err = jsonParser.Decode(&c.eggs)
	} else {
        return fmt.Errorf("Unable to open 'eggs.json' file :: " + err.Error())
    }
    
    if err != nil {
        return fmt.Errorf("eggs.json file appears invalid :: " + err.Error())
    } else if len(c.eggs) < 1 {
        return fmt.Errorf("Please add at least one egg to your nest")
    }
    
    //now that we got our eggs, do a quick check to ensure that the settings are valid
    c.errEggs = make(map[int]bool, 0)
    c.warnEggs = make(map[int]bool, 0)
    for idx, e := range (c.eggs) {
        c.errEggs[idx] = false
        c.warnEggs[idx] = false
        c.validateUrl(e.Url)
        for _, errs := range (e.Errors) {
            c.validateRegex(errs.Regex)
        }
        
        for _, errs := range (e.Warnings) {
            c.validateRegex(errs.Regex)
        }
    }
    
    //now do the crew
    configFile, err = os.Open("crew.json") //try the file
    
	if err == nil {
        defer configFile.Close()
		jsonParser := json.NewDecoder(configFile)
		err = jsonParser.Decode(&c.crew)
	} else {
        return fmt.Errorf("Unable to open 'crew.json' file :: " + err.Error())
    }
    
    if err != nil {
        return fmt.Errorf("crew.json file appears invalid :: " + err.Error())
    } else if len(c.crew) < 1 {
        return fmt.Errorf("Please add at least one crew member")
    }
    
    return nil  //we're good
}

/*! \brief Default entry point.  This will check all the eggs that need to be checked at this interval
 */
func (c *Crow_c) CheckAllEggs () {
    //var err error
    c.elapsedIntervals++    //ramp the interval
    
    for idx, e := range (c.eggs) {    //loop through all the eggs we're supposed to be watching
        if c.elapsedIntervals % e.Interval == 0 {   //this is in our interval window
            err, warn := c.crowUrl.Check(e) //do everything we need to do for this egg
            //fmt.Println("error: ", err, "  warning: ", warn)
            
            if err != nil {
                if c.errEggs[idx] == false {    //make sure we haven't already sent out the message
                    c.errEggs[idx] = true
                    c.warnEggs[idx] = false
                    c.crowSquawk.Squawk(c.crew, e, err, nil)
                }
            } else if warn != nil { //make sure we haven't already sent out the message
                if c.errEggs[idx] == false && c.warnEggs[idx] == false {
                    c.warnEggs[idx] = true
                    c.crowSquawk.Squawk(c.crew, e, nil, warn)
                }
            } else {
                //this egg is healthy, mark it as such
                c.errEggs[idx] = false
                c.warnEggs[idx] = false
            }
        }
    }
}

/*! \brief for sending a message to a specifc crew member
 */
func (c *Crow_c) SendCrewMemberSquawk (alias, message string) (bool) {
    
    for _, cr := range(c.crew) {
        if cr.Alias == alias {
            c.crowSquawk.SendSquawk(cr, message)
            return true
        }
    }
    
    return false
}