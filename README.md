# DotPrompt

[![codecov](https://codecov.io/gh/dazfuller/dotprompt/graph/badge.svg?token=YMLHT2CY3L)](https://codecov.io/gh/dazfuller/dotprompt)

A port of the dotnet [DotPrompt](https://github.com/elastacloud/DotPrompt) library to Go.

The goal is to provide the same functionality and show how the prompt files can be used across languages. The dotnet version utilises the Fluid library which is a dotnet implementation of the [Liquid](https://shopify.github.io/liquid/) templating language. This library makes use of the [Liquid implementation](https://github.com/osteele/liquid) by Oliver Steele.

Fluid contains some methods which are specific to dotnet (such as `format_date`) but where Liquid standard methods are used the templates should be compatible.
