/*! \file url.go
 *  \brief Handles url checks for an egg
 */

package crow

import (
    "fmt"
    "net"
    "net/url"
    "net/http"
    "io/ioutil"
    "regexp"
    "strconv"
    )

type crow_url_c struct {}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Validates the domain of a url we want to check
 *  ie a dns lookup
 */
func (u crow_url_c) domain (inUrl string, retries int) (err error) {
    parts, _ := url.Parse(inUrl)    //ignore the error, we've already checked this
    
    addrs, err := net.LookupHost (parts.Host)
    if err == nil {
        if len(addrs) < 1 {
            err = fmt.Errorf(`Url "%s" didn't resolve to any ip address`, inUrl)
        }
    } else {    //this sometimes times out, to avoid flapping and false positives, i'm doing a retry
        retries--
        if retries > 0 {
            return u.domain(inUrl, retries)
        }
    }
    return
}

/*! \brief Looks for and captures the regexp
 */
func (u crow_url_c) regexCapture (cur capture_t, html string) (error) {
    if len(cur.Regex) > 0 {
		if len(html) > 0 {
			match := regexp.MustCompile(cur.Regex)
			//resp2 := match.FindStringSubmatch(html)
			//fmt.Println(cur.Regex, resp2)
			
			if cur.Exists || cur.Missing {
				resp := match.FindStringIndex(html)
				if cur.Exists && len(resp) != 0 {
					return fmt.Errorf(`%s: exists: %s`, cur.Alias, cur.Regex)
				} else if cur.Missing && len(resp) == 0 {
					return fmt.Errorf(`%s: missing '%s'`, cur.Alias, cur.Regex)
				}
			} else {
				resp := match.FindStringSubmatch(html)
				if len(resp) > 0 {
					if cur.Exists { //this exists and shouldn't
						return fmt.Errorf(`%s: exists: %s`, cur.Alias, cur.Regex)
					}
					
					//see if we can turn this into an int
					val, err := strconv.Atoi(resp[1])
					if err == nil {
						if val > cur.Max {
							return fmt.Errorf(`%s value %d exceeds limit %d`, cur.Alias, val, cur.Max)
						} else if val < cur.Min {
							return fmt.Errorf(`%s value %d below  limit %d`, cur.Alias, val, cur.Min)
						}
					}
				} else {    //we couldn't find it, see if that's bad
					if cur.Missing {
						return fmt.Errorf(`%s: missing '%s'`, cur.Alias, cur.Regex)
					} else if !cur.Exists {
						return fmt.Errorf(`%s: regex error, could not parse, '%s'`, cur.Alias, cur.Regex)
					}
					
				}
			}
		} else {
			return fmt.Errorf("%s: no page body returned", cur.Alias)
		}
    }
    return nil
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Main entry point
 *  This will do all the checks required based on url/http and body parsing
 */
func (u crow_url_c) Check (egg egg_t) (err, warn error) {
    if len(egg.Url) < 4 { return }   //we don't have a url to check
    
    //check the domain
    err = u.domain(egg.Url, 3)  //try 3 times
    if err == nil {
        //now get the page as a request
        resp, err := http.Get(egg.Url)
        if err == nil {
            defer resp.Body.Close()
            body, err := ioutil.ReadAll(resp.Body)
            
            if err == nil {
                html := string(body[:]) //this is our page
                
                //look for errors first, these are worst
                for _, cur := range (egg.Errors) {
                    if err = u.regexCapture (cur, html); err != nil {
                        return err, warn
                    }
                }
                
                //now do the warnings
                for _, cur := range (egg.Warnings) {
                    if warn = u.regexCapture (cur, html); warn != nil {
                        return err, warn
                    }
                }
            } else {
                return err, nil
            }
        } else {
			fmt.Println(resp)
            return err, nil
        }
    }
    return
}