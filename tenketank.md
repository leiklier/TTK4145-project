# Tanker og ideer til tidenes heisprosjekt


- Vi trenger nettverskmodul, fant en på git
- Vi trenger heisdriver (aka et program som kan styre en heis)
- Vi trenger Event handler/control
- Vi trenger Order distributor/cost function  


## Moduler & arbeidsoppgaver

### (1) - Driver

- Representerer en heis. Består av 1 heis og et kontrollpanel. 

## (2) Event handler/control
- Tar i mot en event fra driveren (f.eks trenger heis i 2 etg). 
- Skal utføre handlinger, kjører heisen. Dette er sjåføren, heisen er bilen. Gir beskjed til oRder dist om å finne ut hvilken heis som er nais 

## (3) Order dist/cost function
- Finner ut hvilken heis som skal utføre oppgaven. Sender melding til alle heiser at heis X har fått oppgaven.




## Forberding til design, etter DR:

#### 1: Gjøre ting mindre kopmlisert; forenkle.

- Fjerne Peer to peer. Det var overfladig

- Heis `i` sjekker heis `i+1`. Dette erstatter hashingen. Sjekker betyr at den kontrollerer at heisen utfører instruksene sine innenfor en gitt tid. 

- Ikke overtenke edge-cases. Få på plass basic funksjonalitet. En velfungerende ring. 


#### 2: Hva har blitt gjort?

- Fra øving 4: Kommunikasjon mellom to datamasinker, begge veier. Vi har også broadcast. 

- 

#### 3: Hva må gjøres, i prioritert rekkefølge:

- Network: 
    - Modul for å holde styr på nodene som er tilkoblet
    - Broadcast funksjonalitet
    - Funksjon for å sette opp nettverket
    - Ping
    - Hva skjer når en ny node joiner? Sync state
    - Hva skjer når en node forlater ? redistribuere ordre
    - Funksjon for å kicke en node




## Tanker til cost funksjon

1. Formål: Distrubuere enn hallcall til den mest aktuelle heisen.