# Granti

<p align="center"><img src="https://camo.githubusercontent.com/274385982abc5f3457f94b6b276c619f9e263e24/68747470733a2f2f73746f726167652e676f6f676c65617069732e636f6d2f676f70686572697a656d652e61707073706f742e636f6d2f676f70686572732f376664303165653338333362376536383065363230663064643630323033326330333232386439302e706e67" alt="Granti Gohper" data-canonical-src="https://storage.googleapis.com/gopherizeme.appspot.com/gophers/7fd01ee3833b7e680e620f0dd602032c03228d90.png" height="300px"  style="text-align: center;"></p>

## What is that  
Granti is a tool I've written to check the logs of an application, and, indicating via RegEx the fields of the IP and the timestamp, it analyzes the logs and takes an action when a user exceed with some requests.  
> But Fail2ban does the same thing!  

<img src="https://camo.githubusercontent.com/919a0c825b8fe9cf2cb66c1ffb8d1a46f88bbc83/68747470733a2f2f692e6b796d2d63646e2e636f6d2f656e74726965732f69636f6e732f6f726967696e616c2f3030302f3032382f3539362f64736d47614b574d654858653951754a74715f79733330504e6654476e4d73527548756f5f4d557a4743672e6a7067" alt="Well, yes, but actually no" data-canonical-src="https://i.kym-cdn.com/entries/icons/original/000/028/596/dsmGaKWMeHXe9QuJtq_ys30PNfTGnMsRuHuo_MUzGCg.jpg" height="300px" >
As I saw from my tests, fail2ban looks at the rate of the requests;  
Eg. `I want to ban any user that does more than 100 requests in a 150 seconds time window.`  
For Fail2Ban (from my tests), there is no difference between 100 requests in 150 seconds and 10 requests in 10 seconds.  
This is why I made Granti.    

### How it works
The way Granti works is by abstracting a ring chain structure:  
Every element in the chain rapresents a request made from a certain IP (there is a chain for each IP).
Each element has a number and "contains" the timestamp of a request.  
Given a certain number of chain elements (the maximum request we want to allow an user to do), when the chain "closes" up it overwrite the timestamp of the chain element it's writing to.  
But, before doing so, it check the timestamp of the request.  
If the delta timestamp (between the request that's being overwritten and the request that's going to overwrite) is too low, the IP gets banned (an action gets exectued). 


## Compilation
Set CGO_ENABLED=1 for sqlite  
Command to compile statically and export to a VM:  
CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' .  

## TODO
- [ ] Check for given inputs (eg. makes sure that the numbers are not negative)
- [ ] Create a log file per each jail
- [ ] Create a systemd installer and integration
