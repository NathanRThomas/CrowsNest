# CrowsNest
Site monitoring tool

File based configuration for montitoring websites  
Allows for email and text message alerts based on text capture of the site  

to get this:  
go get github.com/NathanRThomas/CrowsNest  


Email is not in place yet, that's on the to-do list  
Text messages are, and these require a Plivo account right now  
as well as:  
go get github.com/toomore/plivo-go  


This requires three files:  
squawk.json  
crew.json  
eggs.json  

These files can be local to the working directory of the service, or you can set the environment variable 
"CROWSNEST_CONFIG_DIR" to the directory these files are located  

#squawk.json
This contains the information required to send messages.  ie the authentication information
{"plivo":{"number":"18005551234", "auth_id":"AUTHID", "token":"TOKEN"}}  

#crew.json
Contains a list of all members of the crew that we would want to send a notification to  
[{"alias":"NateDogg", "phone":"12075551234", "class_mask":255}]  

alias: Simply for reference  
phone: Full cell number including the 1 and area code  
class_mask: This is used for grouping.  if crew.class_mask & egg.class > 0 { send the message }  


#eggs.json  
Contains a list of servers/urls that we want to monitor.  
Right now we only support Get type requests  
[  
    {"alias":"Api Backend", "url":"https://api.backend.com","interval":1, "class":1,  
    "errors":[{"alias":"This is bad when it's higher", "regex":"\\nUsers: (\\d+)", "max":60},  
            {"alias":"Site Down", "regex":"Things look good!", "missing":true}],  
    "warnings":[{"alias":"This could be bad when it's higher", "regex":"\\nUsers: (\\d+)", "max":30}]  
    }  
]  

alias: Simply for reference, but is included in the notifications about the specific server, so make these unique  
url: What we want to monitor  
interval: In minutes, the frequency to check the url  
class: Again, for grouping.  Referenced against the class_mask for the crew.  if crew.class_mask & egg.class > 0 { send the message }  

errors vs warnings - Warnings can be optional, but allows for a heads up before something bad may happen.  Structure between them is the same  

alias: For reference in the notification about what check went wrong  
regex: A string for whate we're looking for, or capturing  

So, there can be 4 "things" that are wrong based on the current setup  
missing: if this is set to true, then the error state is when this regex expression is missing from the page request. ie "Things look good"  
exists: if this is true then the error state is when this text exists on the page. ie "Database connection error"  

You can also do a max and min.  These need to be interger values.  This requires a capture in the regex.  ie "Users: (\d+)"  
max: if the captured value is GREATER than this int, then we have an error state  
min: if the captured value is LESS than this int, then we have an error state  

#Service
You'll probably want to set this up as a service under systemctl.  Here's an example:  
/etc/systemd/system/crowsnest.service  

[Unit]  
Description=CrowsNest Web server monitor  
After=network.target  
  
[Service]  
WorkingDirectory=/home/NateDogg/crowsnest/  
Type=simple  
User=NateDogg  
ExecStart=/home/NateDogg/crowsnest/CrowsNest  
Restart=on-abort  
  
  
[Install]  
WantedBy=multi-user.target  

Based on this you'd want to have your CrowsNest binary and .json files in the same directory.  In this case /home/NateDogg/crowsnest  

#Testing
This will send a text message to the crew member specified to make sure things are working as expected  
./CrowsNest -testsquawk=NateDogg  

If there is a problem with any of the json files that will be reported immediately upon startup  
journalctl -u crowsnest  
