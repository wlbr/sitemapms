# sitemapms
sitemap.xml generator for websites powered by CMSMS (CMS Made Simple) - www.cmsmadesimple.org

### Usage:
    sitemapms [options] 

sitemapms is a google sitemap generator for websites powered by CMSMS (CMS Made Simple) - www.cmsmadesimple.org. It does not crawl the website itself, but reads the pages and related information straight out of the CMSMS database. Thus it is really fast, its access is not tracked by analytics tools as Piwik etc. and it is perfectly fitting to be run as a cron job.

The commandline options are overruling the configfile settings. So first the configfile gets loaded and afterwards the command line arguments are set.


### Options: 
  &nbsp; -db="": The database name to connect to. <br>
  &nbsp; -ini="$HOME/.sitemapms.ini": The configfile to read parameters from. <br>
  &nbsp; -password="": The database users password. <br>
  &nbsp; -path="./sitemap.xml": The fully qualified pathname to the sitemap incl. filename (e.g.'/var/www/sitemap.xml'). <br>
  &nbsp; -url="": The baseurl of the website (e.g. http://www.google.com) <br>
  &nbsp; -user="": The database user to use for the db connection. <br>


### Example configuration file:

    ; SiteMapMS Config file
    [Database]
      Database = cms_db
      User     = mabuse
      Password = fkhdb4322rb
    [Site]
      BaseUrl     = http://www.mysite.com
      SiteMapPath = /var/www/sitemap.xml
      Filter     = ^/doc/.*
      Filter     = ^/stats.*


## Notes:
   * __Take care__: the tool is rather simple and almost all error handling is omitted.
   * the url priority is calculated by the number of subdirectories of the link
   * url and modification date are taken from the cmsms database
   * urls inside of your CMSMS that are generated dynamically by plugins as Gallery, EventList or News e.g. are skipped, as they are not listed in CMSMS pages table in the database.