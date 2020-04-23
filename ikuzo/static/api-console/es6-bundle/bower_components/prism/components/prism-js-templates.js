(function(Prism){var templateString=Prism.languages.javascript["template-string"],templateLiteralPattern=templateString.pattern.source,interpolationObject=templateString.inside.interpolation,interpolationPunctuationObject=interpolationObject.inside["interpolation-punctuation"],interpolationPattern=interpolationObject.pattern.source;// see the pattern in prism-javascript.js
/**
	 * Creates a new pattern to match a template string with a special tag.
	 *
	 * This will return `undefined` if there is no grammar with the given language id.
	 *
	 * @param {string} language The language id of the embedded language. E.g. `markdown`.
	 * @param {string} tag The regex pattern to match the tag.
	 * @returns {object | undefined}
	 * @example
	 * createTemplate('css', /\bcss/.source);
	 */function createTemplate(language,tag){if(!Prism.languages[language]){return void 0}return{pattern:RegExp("((?:"+tag+")\\s*)"+templateLiteralPattern),lookbehind:!0/* ignoreName */ /* skipSlots */,greedy:!0,inside:{"template-punctuation":{pattern:/^`|`$/,alias:"string"},"embedded-code":{pattern:/[\s\S]+/,alias:language}}}}Prism.languages.javascript["template-string"]=[// styled-jsx:
//   css`a { color: #25F; }`
// styled-components:
//   styled.h1`color: red;`
createTemplate("css",/\b(?:styled(?:\([^)]*\))?(?:\s*\.\s*\w+(?:\([^)]*\))*)*|css(?:\s*\.\s*(?:global|resolve))?|createGlobalStyle|keyframes)/.source),// html`<p></p>`
// div.innerHTML = `<p></p>`
createTemplate("html",/\bhtml|\.\s*(?:inner|outer)HTML\s*\+?=/.source),// svg`<path fill="#fff" d="M55.37 ..."/>`
createTemplate("svg",/\bsvg/.source),// md`# h1`, markdown`## h2`
createTemplate("markdown",/\b(?:md|markdown)/.source),// gql`...`, graphql`...`, graphql.experimental`...`
createTemplate("graphql",/\b(?:gql|graphql(?:\s*\.\s*experimental)?)/.source),// vanilla template string
templateString].filter(Boolean);/**
	 * Returns a specific placeholder literal for the given language.
	 *
	 * @param {number} counter
	 * @param {string} language
	 * @returns {string}
	 */function getPlaceholder(counter,language){return"___"+language.toUpperCase()+"_"+counter+"___"}/**
	 * Returns the tokens of `Prism.tokenize` but also runs the `before-tokenize` and `after-tokenize` hooks.
	 *
	 * @param {string} code
	 * @param {any} grammar
	 * @param {string} language
	 * @returns {(string|Token)[]}
	 */function tokenizeWithHooks(code,grammar,language){var env={code:code,grammar:grammar,language:language};Prism.hooks.run("before-tokenize",env);env.tokens=Prism.tokenize(env.code,env.grammar);Prism.hooks.run("after-tokenize",env);return env.tokens}/**
	 * Returns the token of the given JavaScript interpolation expression.
	 *
	 * @param {string} expression The code of the expression. E.g. `"${42}"`
	 * @returns {Token}
	 */function tokenizeInterpolationExpression(expression){var tempGrammar={"interpolation-punctuation":interpolationPunctuationObject},tokens=Prism.tokenize(expression,tempGrammar);if(3===tokens.length){/**
			 * The token array will look like this
			 * [
			 *     ["interpolation-punctuation", "${"]
			 *     "..." // JavaScript expression of the interpolation
			 *     ["interpolation-punctuation", "}"]
			 * ]
			 */var args=[1,1];args.push.apply(args,tokenizeWithHooks(tokens[1],Prism.languages.javascript,"javascript"));tokens.splice.apply(tokens,args)}return new Prism.Token("interpolation",tokens,interpolationObject.alias,expression)}/**
	 * Tokenizes the given code with support for JavaScript interpolation expressions mixed in.
	 *
	 * This function has 3 phases:
	 *
	 * 1. Replace all JavaScript interpolation expression with a placeholder.
	 *    The placeholder will have the syntax of a identify of the target language.
	 * 2. Tokenize the code with placeholders.
	 * 3. Tokenize the interpolation expressions and re-insert them into the tokenize code.
	 *    The insertion only works if a placeholder hasn't been "ripped apart" meaning that the placeholder has been
	 *    tokenized as two tokens by the grammar of the embedded language.
	 *
	 * @param {string} code
	 * @param {object} grammar
	 * @param {string} language
	 * @returns {Token}
	 */function tokenizeEmbedded(code,grammar,language){// 1. First filter out all interpolations
// because they might be escaped, we need a lookbehind, so we use Prism
/** @type {(Token|string)[]} */var _tokens=Prism.tokenize(code,{interpolation:{pattern:RegExp(interpolationPattern),lookbehind:!0}}),placeholderCounter=0,placeholderMap={},embeddedCode=_tokens.map(function(token){if("string"===typeof token){return token}else{var interpolationExpression=token.content,placeholder;while(-1!==code.indexOf(placeholder=getPlaceholder(placeholderCounter++,language))){}placeholderMap[placeholder]=interpolationExpression;return placeholder}}).join(""),embeddedTokens=tokenizeWithHooks(embeddedCode,grammar,language),placeholders=Object.keys(placeholderMap);// replace all interpolations with a placeholder which is not in the code already
placeholderCounter=0;/**
		 *
		 * @param {(Token|string)[]} tokens
		 * @returns {void}
		 */function walkTokens(tokens){for(var i=0;i<tokens.length;i++){if(placeholderCounter>=placeholders.length){return}var token=tokens[i];if("string"===typeof token||"string"===typeof token.content){var placeholder=placeholders[placeholderCounter],s="string"===typeof token?token:/** @type {string} */token.content,index=s.indexOf(placeholder);if(-1!==index){++placeholderCounter;var before=s.substring(0,index),middle=tokenizeInterpolationExpression(placeholderMap[placeholder]),after=s.substring(index+placeholder.length),replacement=[];if(before){replacement.push(before)}replacement.push(middle);if(after){var afterTokens=[after];walkTokens(afterTokens);replacement.push.apply(replacement,afterTokens)}if("string"===typeof token){tokens.splice.apply(tokens,[i,1].concat(replacement));i+=replacement.length-1}else{token.content=replacement}}}else{var content=token.content;if(Array.isArray(content)){walkTokens(content)}else{walkTokens([content])}}}}walkTokens(embeddedTokens);return new Prism.Token(language,embeddedTokens,"language-"+language,code)}/**
	 * The languages for which JS templating will handle tagged template literals.
	 *
	 * JS templating isn't active for only JavaScript but also related languages like TypeScript, JSX, and TSX.
	 */var supportedLanguages={javascript:!0,js:!0,typescript:!0,ts:!0,jsx:!0,tsx:!0};Prism.hooks.add("after-tokenize",function(env){if(!(env.language in supportedLanguages)){return}/**
		 * Finds and tokenizes all template strings with an embedded languages.
		 *
		 * @param {(Token | string)[]} tokens
		 * @returns {void}
		 */function findTemplateStrings(tokens){for(var i=0,l=tokens.length,token;i<l;i++){token=tokens[i];if("string"===typeof token){continue}var content=token.content;if(!Array.isArray(content)){if("string"!==typeof content){findTemplateStrings([content])}continue}if("template-string"===token.type){/**
					 * A JavaScript template-string token will look like this:
					 *
					 * ["template-string", [
					 *     ["template-punctuation", "`"],
					 *     (
					 *         An array of "string" and "interpolation" tokens. This is the simple string case.
					 *         or
					 *         ["embedded-code", "..."] This is the token containing the embedded code.
					 *                                  It also has an alias which is the language of the embedded code.
					 *     ),
					 *     ["template-punctuation", "`"]
					 * ]]
					 */var embedded=content[1];if(3===content.length&&"string"!==typeof embedded&&"embedded-code"===embedded.type){// get string content
var code=stringContent(embedded),alias=embedded.alias,language=Array.isArray(alias)?alias[0]:alias,grammar=Prism.languages[language];if(!grammar){// the embedded language isn't registered.
continue}content[1]=tokenizeEmbedded(code,grammar,language)}}else{findTemplateStrings(content)}}}findTemplateStrings(env.tokens)});/**
	 * Returns the string content of a token or token stream.
	 *
	 * @param {string | Token | (string | Token)[]} value
	 * @returns {string}
	 */function stringContent(value){if("string"===typeof value){return value}else if(Array.isArray(value)){return value.map(stringContent).join("")}else{return stringContent(value.content)}}})(Prism);