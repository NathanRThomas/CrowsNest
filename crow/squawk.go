/*! \file squawk.go
 *  \brief Handles alerting of the crew about issues
 *  Right now it just handles text messages through Plivo, however can be modified for emails etc
 */

package crow

import (
    "fmt"
    "github.com/NathanRThomas/plivo-go"
    )

type squawk_config_t struct {
    Plivo       struct {
		Number, Token string
		AuthID      string      `json:"auth_id"`
	}
	Slack		struct {
		Username, Token string
	}
	Twilio		struct {
		SID, Token, Number, ShortToken string
	}
}

type crow_squawk_c struct {
    Config      squawk_config_t
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Sends a message to a crew
 */
func (s crow_squawk_c) SendSquawk (c crew_t, message string) {
    //init our account
    account := &plivo.Account{ User: s.Config.Plivo.AuthID, Password: s.Config.Plivo.Token }
    
    if len(c.Phone) > 0 {
        fmt.Printf("Sending text message to %s\n", c.Alias)
        msg := plivo.NewMessage(c.Phone, s.Config.Plivo.Number, message, account)
        msg.Send()
    }
}

/*! \brief Main entry point
 *  This will do all the checks required based on url/http and body parsing
 */
func (s crow_squawk_c) Squawk (crew []crew_t, badEgg egg_t, err, warn error) (error) {
    message := ""
    if err != nil {
        message = fmt.Sprintf("Error on '%s'! %s", badEgg.Alias, err.Error())
    } else if warn != nil {
        message = fmt.Sprintf("Warning on '%s': %s", badEgg.Alias, warn.Error())
    } else {
        return nil  //we shouldn't get here
    }
    
    //fmt.Println(message)    //write this to screen
    
    //now go through and send our messages
    for _, c := range(crew) {
        if c.ClassMask & badEgg.Class > 0 { //this user gets these alerts
            s.SendSquawk (c, message)
        }
    }
    
    return nil
}