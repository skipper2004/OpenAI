import json
import openai
from colorama import Fore, init
from datetime import datetime
init()


class GPT:
    def __init__(self):
        self.data = json.load(open('settings.json'))

        self.api = self.data.get('API')
        self.channels_text = self.data.get('CHANNELS_TEXT')
        self.channels_img = self.data.get('CHANNELS_IMG')

        self.count = -1

    def check(self, error) -> int:
        print(Fore.RED, f'[{datetime.now().time()}] [GET IMG {error}]')

        if str(error) == 'You exceeded your current quota, please check your plan and billing details.':
            self.api.remove(openai.api_key)
            return 0
        return 1

    def get_api(self) -> str:
        self.count += 1

        if self.count == len(self.api):
            self.count = 0

        return self.api[self.count]

    def get_text(self, text) -> str:
        print(Fore.GREEN, f'[{datetime.now().time()}] [GET TEXT: {text[:20]}]')

        while True:
            openai.api_key = self.get_api()

            try:
                return openai.Completion.create(
                    model="text-davinci-003",
                    prompt=text,
                    temperature=0,
                    max_tokens=1000,
                    frequency_penalty=0,
                    presence_penalty=0,
                )['choices'][0]['text']
            except Exception as error:
                print(Fore.RED, f'[{datetime.now().time()}] [GET TEXT {error}]')

                if not self.check(error): continue

                return 'Error'

    def get_img(self, text) -> str:
        print(Fore.GREEN, f'[{datetime.now().time()}] [GET IMG: {text[:20]}]')

        while True:
            openai.api_key = self.get_api()

            try:
                return openai.Image.create(
                    prompt=text,
                    n=1,
                    size="1024x1024"
                )['data'][0]['url']
            except Exception as error:
                print(Fore.RED, f'[{datetime.now().time()}] [GET IMG {error}]')

                if not self.check(error): continue

                return 'Error'
