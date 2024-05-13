package main

const (
	template1 = `我会给你几个问题类型，请参考背景知识（可能为空）和对话记录，判断我“本次问题”的类型，并返回一个问题“类型ID”:
<问题类型>
{"questionType": "website-crawler____getWebsiteContent", "typeId": "wqre"}
{"questionType": "website-crawler____getWeather", "typeId": "sdfa"}
{"questionType": "其他问题", "typeId": "agex"}
</问题类型>

<背景知识>
你将作为系统API协调工具，为我分析给出的content并结合对话记录来判断是否需要执行哪些工具。
工具如下
## Tools
You can use these tools below:
1. [website-crawler____getWebsiteContent] 用于解析网页内容。

2. [website-crawler____getWeather] 用于获取目的地区的天气信息。

##
</背景知识>

<对话记录>
Human:https://github.com/bincooo/edge-api
AI:请问你需要什么帮助？
</对话记录>


content= "{{prompt}}"

类型ID=？
请补充完类型ID=`

	template2 = `你可以从 <对话记录></对话记录> 中提取指定 JSON 信息，你仅需返回 JSON 字符串，无需回答问题。
<提取要求>
{{description}}
</提取要求>

<字段说明>
1. 下面的 JSON 字符串均按照 JSON Schema 的规则描述。
2. key 代表字段名；description 代表字段的描述；required 代表是否必填(true|false)。
3. 如果没有可提取的内容，忽略该字段。
4. 本次需提取的JSON Schema：
{"key":"url", "description":"网页链接", "required": true}
</字段说明>

<对话记录>
Human:https://github.com/bincooo/edge-api
AI:请问你需要什么帮助？
</对话记录>`
)
