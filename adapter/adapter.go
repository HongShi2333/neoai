package adapter

import (
	"neoai/adapter/azure"
	"neoai/adapter/baichuan"
	"neoai/adapter/bing"
	"neoai/adapter/claude"
	adaptercommon "neoai/adapter/common"
	"neoai/adapter/coze"
	"neoai/adapter/dashscope"
	"neoai/adapter/deepseek"
	"neoai/adapter/dify"
	"neoai/adapter/hunyuan"
	"neoai/adapter/midjourney"
	"neoai/adapter/openai"
	"neoai/adapter/palm2"
	"neoai/adapter/skylark"
	"neoai/adapter/slack"
	"neoai/adapter/sparkdesk"
	"neoai/adapter/zhinao"
	"neoai/adapter/zhipuai"
	"neoai/globals"
	"fmt"
)

var channelFactories = map[string]adaptercommon.FactoryCreator{
	globals.OpenAIChannelType:      openai.NewChatInstanceFromConfig,
	globals.AzureOpenAIChannelType: azure.NewChatInstanceFromConfig,
	globals.ClaudeChannelType:      claude.NewChatInstanceFromConfig,
	globals.SlackChannelType:       slack.NewChatInstanceFromConfig,
	globals.BingChannelType:        bing.NewChatInstanceFromConfig,
	globals.PalmChannelType:        palm2.NewChatInstanceFromConfig,
	globals.SparkdeskChannelType:   sparkdesk.NewChatInstanceFromConfig,
	globals.ChatGLMChannelType:     zhipuai.NewChatInstanceFromConfig,
	globals.QwenChannelType:        dashscope.NewChatInstanceFromConfig,
	globals.HunyuanChannelType:     hunyuan.NewChatInstanceFromConfig,
	globals.BaichuanChannelType:    baichuan.NewChatInstanceFromConfig,
	globals.SkylarkChannelType:     skylark.NewChatInstanceFromConfig,
	globals.ZhinaoChannelType:      zhinao.NewChatInstanceFromConfig,
	globals.MidjourneyChannelType:  midjourney.NewChatInstanceFromConfig,
	globals.DeepseekChannelType:    deepseek.NewChatInstanceFromConfig,
	globals.DifyChannelType:        dify.NewChatInstanceFromConfig,
	globals.CozeChannelType:        coze.NewChatInstanceFromConfig,

	globals.MoonshotChannelType: openai.NewChatInstanceFromConfig, // openai format
	globals.GroqChannelType:     openai.NewChatInstanceFromConfig, // openai format
}

func createChatRequest(conf globals.ChannelConfig, props *adaptercommon.ChatProps, hook globals.Hook) error {
	props.Model = conf.GetModelReflect(props.OriginalModel)
	props.Proxy = conf.GetProxy()

	factoryType := conf.GetType()
	if factory, ok := channelFactories[factoryType]; ok {
		return factory(conf).CreateStreamChatRequest(props, hook)
	}

	return fmt.Errorf("unknown channel type %s (channel #%d)", conf.GetType(), conf.GetId())
}

func createVideoRequest(conf globals.ChannelConfig, props *adaptercommon.VideoProps, hook globals.Hook) error {
	props.Model = conf.GetModelReflect(props.OriginalModel)
	props.Proxy = conf.GetProxy()

	factoryType := conf.GetType()
	if creator, ok := channelFactories[factoryType]; ok {
		inst := creator(conf)
		if v, ok := inst.(adaptercommon.VideoFactory); ok {
			return v.CreateVideoRequest(props, hook)
		}
		return fmt.Errorf("video request not supported by channel type %s (channel #%d)", conf.GetType(), conf.GetId())
	}

	return fmt.Errorf("unknown channel type %s (channel #%d)", conf.GetType(), conf.GetId())
}
