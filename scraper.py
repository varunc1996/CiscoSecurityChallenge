import requests
from bs4 import BeautifulSoup

#Function to scrape text from links in given parameter 'links' and 
# print text data to text files in directory 'directory'
#Returns array of strings of embedded links on visited pages
def scrapeText(links, directory, run):
	page = requests.get("http://cvv.club")
	deepLinks = []
	counter = 0
	for i in links:
		print i
		if run == 1:
			getString = "http://" + i
		else:
			getString = i
		fileName = "./" + directory + "/" + i + ".txt"
		if run == 2:
			fileName = "./" + directory + "/" + str(counter) + ".txt"
		counter = counter + 1

		print "GETSTRING:", getString[:7]
		if getString[:7] != "http://":
			print "Got in"
			continue

		try:
			page = requests.get(getString)
		except requests.exceptions.ConnectionError:
			page.status_code = "Connection refused"

		if page.status_code == 200:
			continue

		soup = BeautifulSoup(page.content, "html5lib")
		soupText = soup.get_text().encode('utf-8').strip()

		F = open(fileName, "w")
		F.write(soupText)
		F.close()

		if run == 1:
			for link in soup.find_all('a'):
			    print "FOUND LINK: ", link.get('href')
			    if link.get('href') is not None and len(link.get('href')) > 7:
				    for x in link.get('href'):
				    	if ord(x) > 128:
				    		break
				    deepLinks.append(link.get('href'))
	return deepLinks

#===============================================================================#

#Open text file with URLs and save them into array
fname = "scrapeIDs.txt"
with open(fname) as f:
	content = f.readlines()
content = [x.strip() for x in content]
f.close()

deepScrapes = [] #Will hold all of the links on pages visited

#Scrape text from given URLs and write them into separate text files in
#directory labeled 'extraction' 
deepScrapes = scrapeText(content, "extraction", 1)

print "\n===============================================================\n"
print "DEEP SCRAPES:", deepScrapes

deeperScrapes = scrapeText(deepScrapes, "deepExtraction", 2)