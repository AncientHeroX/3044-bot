package farutils

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/html"

	"github.com/tdewolff/minify/v2"
	mhtml "github.com/tdewolff/minify/v2/html"
)

func getAttr(n *html.Node, key string) (string, bool) {
    for _, attr := range n.Attr {
        if attr.Key == key {
            return attr.Val, true
        }
    }
    return "", false
}

func SubPartSearch(path string, search []string) (string, error) {
    mi := minify.New()

    mi.AddFunc("text/html", mhtml.Minify)

    subPartText, err := ReadHtmlFromFile(path);

    if err != nil {
        return "", err
    }

    minHtml, err := mi.String("text/html", subPartText)

    partdata, err := html.Parse(strings.NewReader(minHtml))

    if err != nil {
        return "", errors.New("Could not Parse")
    }
    
    // content tag
    contag := GetElementsByClass(partdata, "body conbody")
    if len(contag) <= 0 {
        return "", nil
    }

    if search[0] != ""{
        searchResult, err := TextSearch(contag[0], search)

        if err != nil {
            fmt.Printf("Search for %s, in %s Not found", search, path)
            return "", errors.New("No text found")
        }

        return searchResult, nil
    }
    searchResult := GetText(contag[0])

    return searchResult, nil

}

func checkString(str string, searchTerms []string) bool {
    for _, search := range searchTerms {
        if strings.Contains(strings.ToUpper(str), strings.ToUpper(search)) {
            return true
        }
    }
    return false
}
func TextSearch(n *html.Node, search []string) (string, error) {
    var returnText strings.Builder;

    if n.Type == html.TextNode {
        if checkString(n.Data, search) {
            var parentP *html.Node
            for cn := n.Parent; cn != nil; cn = cn.Parent {
                if cn.Data == "p" {
                    parentP = cn
                    break
                }
            }
            if parentP != nil {
                returnText.WriteString(GetText(parentP))

                for cn := parentP.NextSibling; cn != nil && !checkAttr(cn, "class", "ListL1"); cn = cn.NextSibling {
                    returnText.WriteString(GetText(cn))
                }
            }        
        }
    }

    for c := n.FirstChild; c != nil; c = c.NextSibling {
        recText, err := TextSearch(c, search)
        if err == nil {
            returnText.WriteString(fmt.Sprintf("%s\n", recText))
        }
    }

    return strings.Trim(returnText.String(), "\n"), nil
}
func GetPartScope(part int) (string, error) {
    mi := minify.New()

    mi.AddFunc("text/html", mhtml.Minify)

    partScope, err := ReadHtmlFromFile(fmt.Sprintf("src/FARhtml/%d.000.html", part));

    if err != nil {
        return "", err
    }

    minHtml, err := mi.String("text/html", partScope)

    scopedata, err := html.Parse(strings.NewReader(minHtml))

    if err != nil {
        return "", errors.New("Could not Parse")
    }
    
    contag := GetElementsByClass(scopedata, "body conbody")

    scope := GetText(contag[0])

    return scope, nil
}

func GetPartTitle(part int) (string, error) {
    mi := minify.New()

    mi.AddFunc("text/html", mhtml.Minify)

    parthtml, err := ReadHtmlFromFile(fmt.Sprintf("src/FARhtml/Part_%d.html", part));

    if err != nil {
        return "", errors.New("Not a Part")
    }
    minHtml, err := mi.String("text/html", parthtml)

    partdata, err := html.Parse(strings.NewReader(minHtml))

    if err != nil {
        return "", errors.New("Could not Parse")
    }

    titletag := GetElementByID(partdata, "ariaid-title1")

    title := GetText(titletag)

    return title, nil
}

func checkAttr(n *html.Node, attr string, value string) bool {
    if n.Type == html.ElementNode {
        s, ok := getAttr(n, attr)
        if ok && s == value {
            return true
        }
    }
    return false
}


func GetText(n *html.Node) string {
    var text strings.Builder;

    if n.Type == html.TextNode {
        text.WriteString(n.Data)
    }

    for c := n.FirstChild; c != nil; c = c.NextSibling {
        text.WriteString(GetText(c))
    }

    return strings.Trim(text.String(), "\n")
}
func GetElementsByClass(n *html.Node, class string) []*html.Node {
    var elems []*html.Node


    if checkAttr(n, "class", class) {
        elems = append(elems, n)
    }

    for c := n.FirstChild; c != nil; c = c.NextSibling {
        res := GetElementsByClass(c, class)

        if res != nil {
            for _, el := range res {
                elems = append(elems, el)
            }
        }
    }

    return elems
}

func GetElementByID(n *html.Node, id string) *html.Node{
    if checkAttr(n, "id", id) {
        return n
    }

    for c := n.FirstChild; c != nil; c = c.NextSibling {
        res := GetElementByID(c, id)

        if res != nil {
            return res
        }
    }
    return nil
}

func ReadHtmlFromFile(fileName string) (string, error) {

    bs, err := os.ReadFile(fileName)

    if err != nil {
        return "", err
    }

    return string(bs), nil
}

func Parse(text string) (data []string){
    htmlTokenizer := html.NewTokenizer(strings.NewReader(text))

    var vals []string

    for {
        tt := htmlTokenizer.Next()

        switch {
        case tt == html.ErrorToken:
            return vals 
        case tt == html.TextToken:
            vals = append(vals, htmlTokenizer.Token().Data)

        }
    }
}
