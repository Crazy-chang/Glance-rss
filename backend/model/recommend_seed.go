package model

// defaultRecommendSources 内置推荐源（增量 seed，见 SeedRecommendSources）
// 分类区块的展示顺序 = 各分类在此数组中首次出现的先后顺序（见 RecommendCategoryOrder）
// 调整发现页分类顺序：直接移动整个「// —— 分类名 ——」段落即可
// HotRank > 0 进入发现页"热门订阅"横滑区，数字越小越靠前
var defaultRecommendSources = []RecommendSource{
	// —— 科技资讯 ——
	{Title: "36氪", FeedURL: "https://www.36kr.com/feed", SiteURL: "https://36kr.com", Category: "科技资讯", Description: "提供科技与商业领域的资讯、观点和深度报道", HotRank: 1},
	{Title: "爱范儿", FeedURL: "https://www.ifanr.com/feed", SiteURL: "https://www.ifanr.com", Category: "科技资讯", Description: "让未来触手可及，关注消费电子与互联网", HotRank: 2},
	{Title: "少数派", FeedURL: "https://sspai.com/feed", SiteURL: "https://sspai.com", Category: "科技资讯", Description: "高品质数字消费指南，效率工具与工作流", HotRank: 3},
	{Title: "极客公园", FeedURL: "https://www.geekpark.net/rss", SiteURL: "https://www.geekpark.net", Category: "科技资讯", Description: "聚焦科技产品与创新公司的深度报道"},
	{Title: "Solidot", FeedURL: "https://www.solidot.org/index.rss", SiteURL: "https://www.solidot.org", Category: "科技资讯", Description: "奇客资讯，报道科技和互联网相关新闻", HotRank: 10},
	{Title: "InfoQ 中文站", FeedURL: "https://www.infoq.cn/feed", SiteURL: "https://www.infoq.cn", Category: "科技资讯", Description: "面向架构师和技术决策者的技术内容平台", HotRank: 6},
	{Title: "IT之家", FeedURL: "https://www.ithome.com/rss/", SiteURL: "https://www.ithome.com", Category: "科技资讯", Description: "科技新闻与数码产品资讯"},
	{Title: "百度热搜", FeedURL: "https://rss.aishort.top/?type=baidu", SiteURL: "https://top.baidu.com", Category: "科技资讯", Description: "百度搜索实时热点"},
	{Title: "人人都是产品经理", FeedURL: "https://www.woshipm.com/feed", SiteURL: "https://www.woshipm.com", Category: "科技资讯", Description: "产品经理与互联网运营学习社区"},

	// —— 技术博客 ——
	{Title: "阮一峰的网络日志", FeedURL: "https://www.ruanyifeng.com/blog/atom.xml", SiteURL: "https://www.ruanyifeng.com/blog/", Category: "技术博客", Description: "科技爱好者周刊与 Web 技术文章", HotRank: 4},
	{Title: "酷壳 CoolShell", FeedURL: "https://coolshell.cn/feed", SiteURL: "https://coolshell.cn", Category: "技术博客", Description: "陈皓的博客，分享编程与技术思考", HotRank: 7},
	{Title: "美团技术团队", FeedURL: "https://tech.meituan.com/feed/", SiteURL: "https://tech.meituan.com", Category: "技术博客", Description: "美团一线工程实践与技术总结"},
	{Title: "腾讯 AlloyTeam", FeedURL: "http://www.alloyteam.com/feed/", SiteURL: "http://www.alloyteam.com", Category: "技术博客", Description: "腾讯 AlloyTeam 前端团队技术博客"},
	{Title: "张鑫旭-鑫空间", FeedURL: "https://www.zhangxinxu.com/wordpress/feed/", SiteURL: "https://www.zhangxinxu.com", Category: "技术博客", Description: "Web 前端开发，CSS/HTML/JavaScript 心得"},
	{Title: "云风的 BLOG", FeedURL: "https://blog.codingnow.com/atom.xml", SiteURL: "https://blog.codingnow.com", Category: "技术博客", Description: "游戏与底层技术、编程实践随笔"},
	{Title: "月光博客", FeedURL: "https://www.williamlong.info/atom.xml", SiteURL: "https://www.williamlong.info", Category: "技术博客", Description: "关注互联网、搜索引擎和信息技术"},
	{Title: "Thoughtworks 洞见", FeedURL: "https://insights.thoughtworks.cn/feed/", SiteURL: "https://insights.thoughtworks.cn", Category: "技术博客", Description: "软件架构、工程效能与数字化转型洞见"},
	{Title: "全栈应用开发:精益实践", FeedURL: "https://www.phodal.com/blog/feeds/rss/", SiteURL: "https://www.phodal.com", Category: "技术博客", Description: "Phodal 的全栈开发与架构实践博客"},
	{Title: "博客园", FeedURL: "https://www.cnblogs.com/feeds", SiteURL: "https://www.cnblogs.com", Category: "技术博客", Description: "老牌程序员社区，前端精华多"},
	{Title: "CSDN", FeedURL: "https://blog.csdn.net/rss", SiteURL: "https://www.csdn.net", Category: "技术博客", Description: "国内最大技术博客聚合"},
	{Title: "51CTO", FeedURL: "https://www.51cto.com/rss", SiteURL: "https://www.51cto.com", Category: "技术博客", Description: "技术资讯 + 教程"},
	{Title: "腾讯云开发者社区", FeedURL: "https://cloud.tencent.com/developer/rss", SiteURL: "https://cloud.tencent.com/developer", Category: "技术博客", Description: "腾讯云官方技术文章"},
	{Title: "腾讯云专栏", FeedURL: "https://cloud.tencent.com/developer/column/rss", SiteURL: "https://cloud.tencent.com/developer/column", Category: "技术博客", Description: "腾讯云专栏精选"},

	// —— 财经商业 ——
	{Title: "华尔街见闻", FeedURL: "https://dedicated.wallstreetcn.com/rss.xml", SiteURL: "https://wallstreetcn.com", Category: "财经商业", Description: "全球宏观与市场资讯，7x24 财经快讯", HotRank: 8},
	{Title: "虎嗅网", FeedURL: "https://www.huxiu.com/rss/0.xml", SiteURL: "https://www.huxiu.com", Category: "财经商业", Description: "商业资讯与观点平台"},
	{Title: "金十数据", FeedURL: "https://rsshub.dicomp.net/jin10", SiteURL: "https://www.jin10.com", Category: "财经商业", Description: "全球财经快讯与数据日历"},
	{Title: "财联社", FeedURL: "https://rsshub.dicomp.net/cls/depth/1000", SiteURL: "https://www.cls.cn", Category: "财经商业", Description: "7x24 财经快讯与深度报道"},
	{Title: "第一财经 - 新闻", FeedURL: "https://rsshub.dicomp.net/yicai/news", SiteURL: "https://www.yicai.com", Category: "财经商业", Description: "财经新闻与商业报道"},
	{Title: "7*24小时全球财经", FeedURL: "https://rsshub.dicomp.net/10jqka/realtimenews", SiteURL: "https://www.10jqka.com.cn", Category: "财经商业", Description: "同花顺全球财经实时快讯"},
	{Title: "钛媒体", FeedURL: "https://www.tmtpost.com/rss", SiteURL: "https://www.tmtpost.com", Category: "财经商业", Description: "科技与商业深度报道"},
	{Title: "创业邦", FeedURL: "https://rsshub.dicomp.net/cyzone", SiteURL: "https://www.cyzone.cn", Category: "财经商业", Description: "创投与创业资讯"},
	{Title: "财新网", FeedURL: "https://www.caixin.com/rss", SiteURL: "https://www.caixin.com", Category: "财经商业", Description: "财新网深度财经报道"},
	{Title: "福布斯中文", FeedURL: "https://www.forbeschina.com/feed", SiteURL: "https://www.forbeschina.com", Category: "财经商业", Description: "福布斯中国商业资讯"},
	{Title: "财富中文网", FeedURL: "https://plink.anyfeeder.com/fortunechina", SiteURL: "https://www.fortunechina.com", Category: "财经商业", Description: "财富中文网，全球商业财经报道与榜单"},
	{Title: "雪球", FeedURL: "https://plink.anyfeeder.com/weixin/xueqiujinghua", SiteURL: "https://xueqiu.com", Category: "财经商业", Description: "投资交流社区，股市行情与投资者观点"},

	// —— 国内新闻 ——
	{Title: "央视新闻", FeedURL: "https://plink.anyfeeder.com/weixin/cctvnewscenter", SiteURL: "https://news.cctv.com", Category: "国内新闻", Description: "央视新闻，权威国内外时政要闻"},
	{Title: "头条-求是网", FeedURL: "https://plink.anyfeeder.com/qstheory", SiteURL: "http://www.qstheory.cn", Category: "国内新闻", Description: "求是网头条与理论文章"},
	{Title: "新华社新闻_新华网", FeedURL: "https://plink.anyfeeder.com/newscn/whxw", SiteURL: "http://www.news.cn", Category: "国内新闻", Description: "新华社即时新闻报道"},
	{Title: "澎湃新闻", FeedURL: "https://plink.anyfeeder.com/thepaper", SiteURL: "https://www.thepaper.cn", Category: "国内新闻", Description: "时政与社会深度报道"},
	{Title: "新浪体育", FeedURL: "https://plink.anyfeeder.com/weixin/sports_sina", SiteURL: "https://sports.sina.com.cn", Category: "国内新闻", Description: "新浪体育，国内外体育赛事资讯"},

	// —— 国际新闻 ——
	{Title: "BBC 中文", FeedURL: "https://feeds.bbci.co.uk/zhongwen/simp/rss.xml", SiteURL: "https://www.bbc.com/zhongwen/simp", Category: "国际新闻", Description: "BBC 中文网国际与中国相关报道"},
	{Title: "纽约时报中文网", FeedURL: "https://cn.nytimes.com/rss.html", SiteURL: "https://cn.nytimes.com", Category: "国际新闻", Description: "纽约时报中文国际新闻与深度分析"},
	{Title: "德国之声中文网", FeedURL: "https://rss.dw.com/rdf/rss-chi-all", SiteURL: "https://www.dw.com/zh", Category: "国际新闻", Description: "德国之声中文国际资讯与专题报道"},

	// —— AI 人工智能 ——
	{Title: "机器之心", FeedURL: "https://www.jiqizhixin.com/rss", SiteURL: "https://www.jiqizhixin.com", Category: "AI 人工智能", Description: "国内最大 AI 媒体，论文解读 + 产业动态", HotRank: 5},
	{Title: "Rundown AI", FeedURL: "https://www.therundown.ai/feed", SiteURL: "https://www.therundown.ai", Category: "AI 人工智能", Description: "每天一封 AI 要闻简报"},
	{Title: "OpenAI Blog", FeedURL: "https://openai.com/blog/rss.xml", SiteURL: "https://openai.com/blog", Category: "AI 人工智能", Description: "OpenAI 官方博客，模型发布公告与研究文章"},
	{Title: "arXiv cs.AI", FeedURL: "https://rss.arxiv.org/rss/cs.AI", SiteURL: "https://arxiv.org/list/cs.AI/recent", Category: "AI 人工智能", Description: "AI 方向每日新预印本直出"},
	{Title: "SemiAnalysis", FeedURL: "https://semianalysis.com/feed", SiteURL: "https://semianalysis.com", Category: "AI 人工智能", Description: "LLM 推理成本与芯片深度分析"},
	{Title: "LangChain Blog", FeedURL: "https://blog.langchain.dev/rss/", SiteURL: "https://blog.langchain.dev", Category: "AI 人工智能", Description: "LangChain / LangGraph 官方更新"},

	// —— 设计艺术 ——
	{Title: "优设网", FeedURL: "https://www.uisdc.com/feed", SiteURL: "https://www.uisdc.com", Category: "设计艺术", Description: "国内最大 UI/UX 设计门户"},
	{Title: "腾讯 ISUX", FeedURL: "https://isux.tencent.com/feed", SiteURL: "https://isux.tencent.com", Category: "设计艺术", Description: "腾讯社交产品用户体验设计团队"},

	// —— 数码生活 ——
	{Title: "数字尾巴", FeedURL: "https://www.dgtle.com/feed", SiteURL: "https://www.dgtle.com", Category: "数码生活", Description: "分享美好数字生活，数码评测与体验"},
	{Title: "什么值得买", FeedURL: "https://feed.smzdm.com/", SiteURL: "https://www.smzdm.com", Category: "数码生活", Description: "消费决策门户，好价与商品评测"},
}
