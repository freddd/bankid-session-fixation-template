# bankid-session-fixation-template
A quick and dirty template for automating bankid session fixation outlined in https://mastersplinter.work/research/bankid/

## Usage
This template needs to be adjusted to each target by:
1) Adding url(s)
2) Make the headless browser click & navigate to the correct place
3) Reading the JS response and extract the `autostarttoken` (or extract it from the DOM)

