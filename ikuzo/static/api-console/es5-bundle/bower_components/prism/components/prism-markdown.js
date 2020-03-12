(function(Prism){// Allow only one line break
var inner=/(?:\\.|[^\\\n\r]|(?:\r?\n|\r)(?!\r?\n|\r))/.source;/**
	 * This function is intended for the creation of the bold or italic pattern.
	 *
	 * This also adds a lookbehind group to the given pattern to ensure that the pattern is not backslash-escaped.
	 *
	 * _Note:_ Keep in mind that this adds a capturing group.
	 *
	 * @param {string} pattern
	 * @param {boolean} starAlternative Whether to also add an alternative where all `_`s are replaced with `*`s.
	 * @returns {RegExp}
	 */function createInline(pattern,starAlternative){pattern=pattern.replace(/<inner>/g,inner);if(starAlternative){pattern=pattern+"|"+pattern.replace(/_/g,"\\*")}return RegExp(/((?:^|[^\\])(?:\\{2})*)/.source+"(?:"+pattern+")")}var tableCell=/(?:\\.|``.+?``|`[^`\r\n]+`|[^\\|\r\n`])+/.source,tableRow=/\|?__(?:\|__)+\|?(?:(?:\r?\n|\r)|$)/.source.replace(/__/g,tableCell),tableLine=/\|?[ \t]*:?-{3,}:?[ \t]*(?:\|[ \t]*:?-{3,}:?[ \t]*)+\|?(?:\r?\n|\r)/.source;Prism.languages.markdown=Prism.languages.extend("markup",{});Prism.languages.insertBefore("markdown","prolog",{blockquote:{// > ...
pattern:/^>(?:[\t ]*>)*/m,alias:"punctuation"},table:{pattern:RegExp("^"+tableRow+tableLine+"(?:"+tableRow+")*","m"),inside:{"table-data-rows":{pattern:RegExp("^("+tableRow+tableLine+")(?:"+tableRow+")*$"),lookbehind:!0,inside:{"table-data":{pattern:RegExp(tableCell),inside:Prism.languages.markdown},punctuation:/\|/}},"table-line":{pattern:RegExp("^("+tableRow+")"+tableLine+"$"),lookbehind:!0,inside:{punctuation:/\||:?-{3,}:?/}},"table-header-row":{pattern:RegExp("^"+tableRow+"$"),inside:{"table-header":{pattern:RegExp(tableCell),alias:"important",inside:Prism.languages.markdown},punctuation:/\|/}}}},code:[{// Prefixed by 4 spaces or 1 tab and preceded by an empty line
pattern:/(^[ \t]*(?:\r?\n|\r))(?: {4}|\t).+(?:(?:\r?\n|\r)(?: {4}|\t).+)*/m,lookbehind:!0,alias:"keyword"},{// `code`
// ``code``
pattern:/``.+?``|`[^`\r\n]+`/,alias:"keyword"},{// ```optional language
// code block
// ```
pattern:/^```[\s\S]*?^```$/m,greedy:!0,inside:{"code-block":{pattern:/^(```.*(?:\r?\n|\r))[\s\S]+?(?=(?:\r?\n|\r)^```$)/m,lookbehind:!0},"code-language":{pattern:/^(```).+/,lookbehind:!0},punctuation:/```/}}],title:[{// title 1
// =======
// title 2
// -------
pattern:/\S.*(?:\r?\n|\r)(?:==+|--+)(?=[ \t]*$)/m,alias:"important",inside:{punctuation:/==+$|--+$/}},{// # title 1
// ###### title 6
pattern:/(^\s*)#+.+/m,lookbehind:!0,alias:"important",inside:{punctuation:/^#+|#+$/}}],hr:{// ***
// ---
// * * *
// -----------
pattern:/(^\s*)([*-])(?:[\t ]*\2){2,}(?=\s*$)/m,lookbehind:!0,alias:"punctuation"},list:{// * item
// + item
// - item
// 1. item
pattern:/(^\s*)(?:[*+-]|\d+\.)(?=[\t ].)/m,lookbehind:!0,alias:"punctuation"},"url-reference":{// [id]: http://example.com "Optional title"
// [id]: http://example.com 'Optional title'
// [id]: http://example.com (Optional title)
// [id]: <http://example.com> "Optional title"
pattern:/!?\[[^\]]+\]:[\t ]+(?:\S+|<(?:\\.|[^>\\])+>)(?:[\t ]+(?:"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|\((?:\\.|[^)\\])*\)))?/,inside:{variable:{pattern:/^(!?\[)[^\]]+/,lookbehind:!0},string:/(?:"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|\((?:\\.|[^)\\])*\))$/,punctuation:/^[\[\]!:]|[<>]/},alias:"url"},bold:{// **strong**
// __strong__
// allow one nested instance of italic text using the same delimiter
pattern:createInline(/__(?:(?!_)<inner>|_(?:(?!_)<inner>)+_)+__/.source,!0),lookbehind:!0,greedy:!0,inside:{content:{pattern:/(^..)[\s\S]+(?=..$)/,lookbehind:!0,inside:{}// see below
},punctuation:/\*\*|__/}},italic:{// *em*
// _em_
// allow one nested instance of bold text using the same delimiter
pattern:createInline(/_(?:(?!_)<inner>|__(?:(?!_)<inner>)+__)+_/.source,!0),lookbehind:!0,greedy:!0,inside:{content:{pattern:/(^.)[\s\S]+(?=.$)/,lookbehind:!0,inside:{}// see below
},punctuation:/[*_]/}},strike:{// ~~strike through~~
// ~strike~
pattern:createInline(/(~~?)(?:(?!~)<inner>)+?\2/.source,!1),lookbehind:!0,greedy:!0,inside:{content:{pattern:/(^~~?)[\s\S]+(?=\1$)/,lookbehind:!0,inside:{}// see below
},punctuation:/~~?/}},url:{// [example](http://example.com "Optional title")
// [example][id]
// [example] [id]
pattern:createInline(/!?\[(?:(?!\])<inner>)+\](?:\([^\s)]+(?:[\t ]+"(?:\\.|[^"\\])*")?\)| ?\[(?:(?!\])<inner>)+\])/.source,!1),lookbehind:!0,greedy:!0,inside:{variable:{pattern:/(\[)[^\]]+(?=\]$)/,lookbehind:!0},content:{pattern:/(^!?\[)[^\]]+(?=\])/,lookbehind:!0,inside:{}// see below
},string:{pattern:/"(?:\\.|[^"\\])*"(?=\)$)/}}}});["url","bold","italic","strike"].forEach(function(token){["url","bold","italic","strike"].forEach(function(inside){if(token!==inside){Prism.languages.markdown[token].inside.content.inside[inside]=Prism.languages.markdown[inside]}})});Prism.hooks.add("after-tokenize",function(env){if("markdown"!==env.language&&"md"!==env.language){return}function walkTokens(tokens){if(!tokens||"string"===typeof tokens){return}for(var i=0,l=tokens.length,token;i<l;i++){token=tokens[i];if("code"!==token.type){walkTokens(token.content);continue}/*
				 * Add the correct `language-xxxx` class to this code block. Keep in mind that the `code-language` token
				 * is optional. But the grammar is defined so that there is only one case we have to handle:
				 *
				 * token.content = [
				 *     <span class="punctuation">```</span>,
				 *     <span class="code-language">xxxx</span>,
				 *     '\n', // exactly one new lines (\r or \n or \r\n)
				 *     <span class="code-block">...</span>,
				 *     '\n', // exactly one new lines again
				 *     <span class="punctuation">```</span>
				 * ];
				 */var codeLang=token.content[1],codeBlock=token.content[3];if(codeLang&&codeBlock&&"code-language"===codeLang.type&&"code-block"===codeBlock.type&&"string"===typeof codeLang.content){// this might be a language that Prism does not support
var alias="language-"+codeLang.content.trim().split(/\s+/)[0].toLowerCase();// add alias
if(!codeBlock.alias){codeBlock.alias=[alias]}else if("string"===typeof codeBlock.alias){codeBlock.alias=[codeBlock.alias,alias]}else{codeBlock.alias.push(alias)}}}}walkTokens(env.tokens)});Prism.hooks.add("wrap",function(env){if("code-block"!==env.type){return}for(var codeLang="",i=0,l=env.classes.length;i<l;i++){var cls=env.classes[i],match=/language-(.+)/.exec(cls);if(match){codeLang=match[1];break}}var grammar=Prism.languages[codeLang];if(!grammar){if(codeLang&&"none"!==codeLang&&Prism.plugins.autoloader){var id="md-"+new Date().valueOf()+"-"+Math.floor(1e16*Math.random());env.attributes.id=id;Prism.plugins.autoloader.loadLanguages(codeLang,function(){var ele=document.getElementById(id);if(ele){ele.innerHTML=Prism.highlight(ele.textContent,Prism.languages[codeLang],codeLang)}})}}else{// reverse Prism.util.encode
var code=env.content.replace(/&lt;/g,"<").replace(/&amp;/g,"&");env.content=Prism.highlight(code,grammar,codeLang)}});Prism.languages.md=Prism.languages.markdown})(Prism);