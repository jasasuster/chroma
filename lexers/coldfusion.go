package lexers

import (
	. "github.com/alecthomas/chroma/v2" // nolint
)

var CFStatement = Register(MustNewLexer(
	&Config{
		Name:      "CFStatement",
		Aliases:   []string{"cfs"},
		Filenames: []string{},
		MimeTypes: []string{},
		CaseInsensitive: true,
	},
	cfstatementRules,
))

/*
 * Coldfusion statements
 */
func cfstatementRules() Rules {
	return Rules{
		"root": {
			{`//.*?\n`, CommentSingle, nil},
			{`/\*(?:.|\n)*?\*/`, CommentMultiline, nil},
			{`\+\+|--`, Operator, nil},
			{`<=|>=|==|\+\+|--|\|\||&&`, Operator, nil},
			{`[-+*/^&=!<>?]`, Operator, nil},
			{`mod\b`, Operator, nil},
			{`(eq|lt|gt|lte|gte|not|is|and|or)\b`, Operator, nil},
			{`\|\||&&`, Operator, nil},
			{`\?`, Operator, nil},
			{`"`, StringDouble, Push("string")},
			{`'.*?'`, StringSingle, nil},
			{`\d+`, Number, nil},
			{
	    Words(`\b`, `\b`,
	        "if",
	        "else",
	        "len",
	        "var",
	        "xml",
	        "default",
	        "break",
	        "switch",
	        "component",
	        "property",
	        "function",
	        "do",
	        "try",
	        "catch",
	        "in",
	        "continue",
	        "for",
	        "return",
	        "while",
	        "required",
	        "any",
	        "array",
	        "binary",
	        "boolean",
	        "date",
	        "guid",
	        "numeric",
	        "query",
	        "string",
	        "struct",
	        "uuid",
	        "case",
		    ),
		    Keyword,
		    nil,
			},
			{`(application|session|client|cookie|super|this|variables|arguments)\b`, NameConstant, nil},
			{`([a-z_$][\w.]*)(\s*)(\()`,
				ByGroups(NameFunction, Text, Punctuation), nil},
			{`[a-z_$][\w.]*`, NameVariable, nil},
			{`[()\[\]{};:,.\\]`, Punctuation, nil},
			{`\s+`, Text, nil},
			{`.`, Other, nil},
		},
		"string": {
			{`""`, StringDouble, nil},
			{`#.+?#`, StringInterpol, nil},
			{`[^"#]+`, StringDouble, nil},
			{`#`, StringDouble, nil},
			{`"`, StringDouble, Pop(1)},
		},
	}
}

/*
 * Coldfusion markup only
 */
var Coldfusion = Register(MustNewLexer(
  &Config{
    Name:      "Coldfusion",
    Aliases:   []string{"cf"},
    Filenames: []string{"*.cfm", "*.cfml"},
    MimeTypes: []string{"text/x-coldfusion"},
  },
  coldFusionRules,
))

func coldFusionRules() Rules {
  return Rules{
    "root": {
      {`[^<]+`, Other, nil},
      Include("tags"),
      {`<[^<>]*`, Other, nil},
    },

    "tags": {
      // CF comments
      {`<!---`, CommentMultiline, Push("cfcomment")},
      // HTML comments
      {`(?s)<!--.*?-->`, Comment, nil},
      // cfoutput
      {`(?i)<cfoutput.*?>`, NameTag, Push("cfoutput")},
      // cfscript blocks
      {`(?is)(<cfscript.*?>)(.+?)(</cfscript.*?>)`, ByGroups(NameTag, Using("CFStatement"), NameTag), nil},
      // Generic CF tags
      {`(?is)(</?cf(?:component|include|if|else|elseif|loop|return|dbinfo|dump|abort|location|invoke|throw|file|savecontent|mailpart|mail|header|content|zip|image|lock|argument|try|catch|break|directory|http|set|function|param|silent|processingdirective|setting|query|queryparam)\b)(.*?)(>)`, ByGroups(NameTag, Using("CFStatement"), NameTag), nil},
      {`(?is)(<cfquery\b.*?>)(.*?)(</cfquery\b.*?>)`, ByGroups(NameTag, Using("Coldfusion SQL"), NameTag), nil},
    },

    "cfoutput": {
      {`[^#<]+`, Other, nil},
      {`(#)(.*?)(#)`, ByGroups(Punctuation, Using("CFStatement"), Punctuation), nil},
      {`(?i)</cfoutput.*?>`, NameTag, Pop(1)},
      Include("tags"),
      {`(?s)<[^<>]*`, Other, nil},
      {`#`, Other, nil},
    },
    "cfcomment": {
      {`<!---`, CommentMultiline, Push()},
      {`--->`, CommentMultiline, Pop(1)},
      {`([^<-]|<(?!!---)|-(?!-->))+`, CommentMultiline, nil},
    },
  }
}

/*
 * Coldfusion markup in html
 */
var ColdFusionHtml = Register(DelegatingLexer(HTML, MustNewLexer(
	&Config{
		Name:      "Coldfusion HTML",
		Aliases:   []string{"cfm"},
		Filenames: []string{"*.cfm", "*.cfml"},
		MimeTypes: []string{"application/x-coldfusion"},
	},
	coldFusionRules,
)));

/*
 * Coldfusion markup/script components
 */
var coldFusionCfc = Register(DelegatingLexer(ColdFusionHtml, MustNewLexer(
	&Config{
		Name:      "Coldfusion CFC",
		Aliases:   []string{"cfc"},
		Filenames: []string{"*.cfc"},
		MimeTypes: []string{},
	},
	cfstatementRules,
)));

var ColdFusionSQL = Register(DelegatingLexer(Get("sql"), MustNewLexer(
  &Config{
    Name:      "Coldfusion SQL",
    Aliases:   []string{"cfsql"},
    Filenames: []string{},
    MimeTypes: []string{},
  },
  coldFusionSQLRules,
)));

func coldFusionSQLRules() Rules {
  return Rules{
    "root": {
      // CF comments inside SQL
      {`<!---`, CommentMultiline, Push("cfcomment")},

      // CF tags inside SQL, especially <cfqueryparam>
      {`(?is)(</?cf[a-z0-9_-]+\b)(.*?)(>)`, ByGroups(NameTag, Using("CFStatement"), NameTag), nil},

      // #expression# interpolation inside SQL
      {`(#)(.*?)(#)`, ByGroups(Punctuation, Using("CFStatement"), Punctuation), nil},

      // Everything else should be delegated to SQL
      {`[^<#]+`, Other, nil},
      {`[<#]`, Other, nil},
    },

    "cfcomment": {
      {`<!---`, CommentMultiline, Push()},
      {`--->`, CommentMultiline, Pop(1)},
      {`([^<-]|<(?!!---)|-(?!-->))+`, CommentMultiline, nil},
    },
  }
}
