package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

import _ "github.com/go-sql-driver/mysql"
import "code.google.com/p/gcfg"

type Config struct {
	DataBase struct {
		Database string
		User     string
		Password string
	}
	Site struct {
		BaseUrl     string
		SiteMapPath string   // fully qualified pathname incl. filename
		Filter      []string // any url that matches any of these regex will be suppressed
	}
	ConfigFile string
}

type Flags struct {
	ConfigFile  string
	Database    string
	User        string
	Password    string
	BaseUrl     string
	SiteMapPath string
}

func commandLineParsing(flags *Flags) {
	flag.StringVar(&flags.ConfigFile, "ini", "$HOME/.sitemapms.ini", "The configfile to read parameters from.")
	flag.StringVar(&flags.Database, "db", "", "The database name to connect to.")
	flag.StringVar(&flags.User, "user", "", "The database user to use for the db connection.")
	flag.StringVar(&flags.Password, "password", "", "The database users password.")
	flag.StringVar(&flags.BaseUrl, "url", "", "The baseurl of the website (e.g. http://www.google.com),")
	flag.StringVar(&flags.SiteMapPath, "path", "./sitemap.xml", "The fully qualified pathname to the sitemap incl. filename (e.g.'/var/www/sitemap.xml').")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n"+
			"  %s [options] \n\n"+
			"%s is a google sitemap generator for websites powered by CMSMS (CMS Made Simple) - www.cmsmadesimple.org "+
			"It does not crawl the website itself, but reads the pages and related information straight out of the "+
			"CMSMS database. Thus it is a really fast, its access is not tracked by analytics tools as Piwik etc. "+
			"and perfectly fitting to be run as a cron job.\n\n"+
			"The commandline options are overruling the configfile settings. So first the configfile gets loaded "+
			"and afterwards the command line arguments are set.\n\n"+
			"Example configuration file:\n<snip>\n "+
			"   ; SiteMapMS Config file\n"+
			"   [Database]\n"+
			"     Database = cms_db\n"+
			"     User     = mabuse\n"+
			"     Password = fkhdb4322rb\n"+
			"   [Site]\n"+
			"     BaseUrl     = http://www.mysite.com\n"+
			"     SiteMapPath = /var/www/sitemap.xml\n"+
			"     Filter     = ^/doc/.*\n"+
			"     Filter     = ^/stats.*\n</snip>\n\n"+
			"Options:\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func readConfig(cfg *Config) {

	var flags = new(Flags)
	commandLineParsing(flags)

	cfg.ConfigFile = os.ExpandEnv(flags.ConfigFile)
	err := gcfg.ReadFileInto(cfg, cfg.ConfigFile)
	if err != nil {
		panic(err.Error()) // No ErrorHandling so far
	}

	if flags.Database != "" {
		cfg.DataBase.Database = flags.Database
	}
	if flags.User != "" {
		cfg.DataBase.User = flags.User
	}
	if flags.Password != "" {
		cfg.DataBase.Password = flags.Password
	}
	if flags.BaseUrl != "" {
		cfg.Site.BaseUrl = flags.BaseUrl
	}
	if flags.SiteMapPath != "./sitemap.xml" || cfg.Site.SiteMapPath == "" {
		cfg.Site.SiteMapPath = flags.SiteMapPath
	}
	cfg.Site.SiteMapPath = os.ExpandEnv(cfg.Site.SiteMapPath)
}

func formatItem(cfg *Config, title string, path string, modifydate string) string {

	//2015-01-07 01:32:39 --> 2015-01-08T18:04:27+00:00
	rex, _ := regexp.Compile("(.+) (.+)")
	mod := rex.ReplaceAllString(modifydate, "${1}T$2+00:00")

	u := strings.TrimRight(cfg.Site.BaseUrl, "/")

	prio := 1 / float32(strings.Count(path, "/")+1)

	return fmt.Sprintf(" <url>\n   <loc>%s/%s</loc>\n   <lastmod>%s</lastmod>\n   <changefreq>daily</changefreq>\n   <priority>%.2f</priority>\n </url>\n",
		u, path, mod, prio)
}

func main() {

	var config = new(Config)
	var title, path, modifydate string

	readConfig(config)

	fmt.Printf("ConfigFile: %s\n", config.ConfigFile)
	fmt.Printf("Database: %s\n", config.DataBase.Database)
	fmt.Printf("User: %s\n", config.DataBase.User)
	fmt.Printf("Password: ***\n") //sorry
	fmt.Printf("BaseUrl: %s\n", config.Site.BaseUrl)
	fmt.Printf("SiteMapPath: %s\n", config.Site.SiteMapPath)
	fmt.Printf("Filters: %s\n", config.Site.Filter)

	//create the filter regular expressions
	rexes := make([]*regexp.Regexp, len(config.Site.Filter))
	for i, filter := range config.Site.Filter {
		rexes[i], _ = regexp.Compile(filter)
	}

	//db, err := sql.Open("mysql", "user:password@/dbname")
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", config.DataBase.User, config.DataBase.Password, config.DataBase.Database))
	if err != nil {
		panic(err.Error()) // No ErrorHandling so far
	}
	defer db.Close()

	rows, err := db.Query("select content_name, hierarchy_path, modified_date from cms_content where type = 'content' and active = 1;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	//create file
	var out *os.File
	if out, err = os.Create(config.Site.SiteMapPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating %s: %v", config.Site.SiteMapPath, err)
		os.Exit(-1)
	}
	defer out.Close()

	fmt.Fprintln(out, `<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintln(out, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">`)

	for rows.Next() {
		err := rows.Scan(&title, &path, &modifydate)
		if err != nil {
			log.Fatal(err)
		}
		show := true
		for _, rex := range rexes {
			if rex.MatchString(path) {
				show = false
			}
		}
		if show {
			fmt.Fprintf(out, formatItem(config, title, path, modifydate))
		}

	}
	fmt.Fprintf(out, "</urlset>")
}
