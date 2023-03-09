import json
import discord
from discord.ext import commands
import asyncio
from gpt import GPT


token = json.load(open('settings.json')).get('TOKEN_BOT')
bot_id = json.load(open('settings.json')).get('ID_BOT')

gpt = GPT()


def main():
    intents = discord.Intents.all()
    client = commands.Bot(command_prefix='!', intents=intents)

    @client.event
    async def on_message(ctx):
        if str(ctx.author.id) != bot_id:
            if str(ctx.channel.id) in gpt.channels_text:
                text = ctx.content

                loop = asyncio.get_event_loop()
                resp = loop.run_in_executor(None, gpt.get_text, text)
                resp = await resp

                await ctx.reply(resp)

            if str(ctx.channel.id) in gpt.channels_img:
                text = ctx.content

                loop = asyncio.get_event_loop()
                resp = loop.run_in_executor(None, gpt.get_img, text)
                resp = await resp

                await ctx.reply(resp)

        await client.process_commands(ctx)

    client.run(token)


if __name__ == '__main__':
    main()