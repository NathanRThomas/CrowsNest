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
    Threshold       int     `json:"threshold"`
    Errors          []capture_t  `json:"errors"`
    Warnings        []capture_t  `json:"warnings"`
}

type crew_t struct {
    Alias           string  `json:"alias"`
    Phone           string  `json:"phone"`
    Email           string  `json:"email"`
    ClassMask       int     `json:"class_mask"`
}

type eggState_t struct {
    err, warn   bool
    errCnt, warnCnt int
}

type Crow_c struct {
    eggs                []egg_t
    crew                []crew_t
    elapsedIntervals    int
    eggStates           []eggState_t
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
    //see if the user set a environment varible for the path location
    configPath := os.Getenv("CROWSNEST_CONFIG_DIR")
    
    //Read in the eggs
    configFile, err := os.Open(configPath + "eggs.json") //try the file
    
	if err == nil {
        defer configFile.Close()
		jsonParser := json.NewDecoder(configFile)
		err = jsonParser.Decode(&c.eggs)
	} else {
        return fmt.Errorf("Unable to open '%seggs.json' file :: " + err.Error(), configPath)
    }
    
    if err != nil {
        return fmt.Errorf("%seggs.json file appears invalid :: " + err.Error(), configPath)
    } else if len(c.eggs) < 1 {
        return fmt.Errorf("Please add at least one egg to your nest")
    }
    
    //now that we got our eggs, do a quick check to ensure that the settings are valid
    c.eggStates = make([]eggState_t, len(c.eggs))   //init our states array, keeps track of the eggs locally
    for _, e := range (c.eggs) {
        c.validateUrl(e.Url)
        for _, errs := range (e.Errors) {
            c.validateRegex(errs.Regex)
        }
        
        for _, errs := range (e.Warnings) {
            c.validateRegex(errs.Regex)
        }
    }
    
    //now do the crew
    configFile, err = os.Open(configPath + "crew.json") //try the file
    
	if err == nil {
        defer configFile.Close()
		jsonParser := json.NewDecoder(configFile)
		err = jsonParser.Decode(&c.crew)
	} else {
        return fmt.Errorf("Unable to open '%screw.json' file :: " + err.Error(), configPath)
    }
    
    if err != nil {
        return fmt.Errorf("%screw.json file appears invalid :: " + err.Error(), configPath)
    } else if len(c.crew) < 1 {
        return fmt.Errorf("Please add at least one crew member")
    }
    
    //now do the squawk file
    configFile, err = os.Open(configPath + "squawk.json") //try the file
    
	if err == nil {
        defer configFile.Close()
		jsonParser := json.NewDecoder(configFile)
		err = jsonParser.Decode(&c.crowSquawk.Config)
	} else {
        return fmt.Errorf("Unable to open '%ssquawk.json' file :: " + err.Error(), configPath)
    }
    
    if err != nil {
        return fmt.Errorf("%ssquawk.json file appears invalid :: " + err.Error(), configPath)
    } else if len(c.crowSquawk.Config.Plivo.AuthID) < 1 {
        return fmt.Errorf("Plivo auth_id is not set")
    } else if len(c.crowSquawk.Config.Plivo.Number) < 1 {
        return fmt.Errorf("Plivo number is not set")
    } else if len(c.crowSquawk.Config.Plivo.Token) < 1 {
        return fmt.Errorf("Plivo token is not set")
    }
    
    return nil  //we're good
}

/*! \brief Default entry point.  This will check all the eggs that need to be checked at this interval
 *I'm keeping seperate err and warn messages, so we can track that something has gone from a warning state, to an error
 *state and that we need to send a new message to alert the user that the state has gotten worse
 */
func (c *Crow_c) CheckAllEggs () {
    //var err error
    c.elapsedIntervals++    //ramp the interval, this is used as a global counter
    
    for idx, e := range (c.eggs) {    //loop through all the eggs we're supposed to be watching
        if c.elapsedIntervals % e.Interval == 0 {   //this is in our interval window for this specific egg
            err, warn := c.crowUrl.Check(e) //do everything we need to do for this egg
            //fmt.Println("error: ", err, "  warning: ", warn)
            
            if err != nil {
                if c.eggStates[idx].err == false {    //make sure we haven't already sent out the message
                    c.eggStates[idx].errCnt += 1    //ramp the error count
                    if c.eggStates[idx].errCnt >= e.Threshold { //check to see if we're past our threshold for this error
                        c.eggStates[idx].err = true
                        c.eggStates[idx].warn = false
                        c.crowSquawk.Squawk(c.crew, e, err, nil)
                    }
                }
            } else if warn != nil { //make sure we haven't already sent out the message
                if c.eggStates[idx].err == false && c.eggStates[idx].warn == false {
                    c.eggStates[idx].warnCnt += 1    //ramp the warning count
                    if c.eggStates[idx].warnCnt >= e.Threshold { //check to see if we're past our threshold for this error
                        c.eggStates[idx].warn = true
                        c.crowSquawk.Squawk(c.crew, e, nil, warn)
                    }
                }
            } else {
                //this egg is healthy, mark it as such
                c.eggStates[idx] = eggState_t{} //empty struct defaults it
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