from office365.runtime.auth.user_credential import UserCredential
from office365.sharepoint.client_context import ClientContext
import os, tempfile,sys

settings = {
    'username': os.environ["o365-user"],
    'password': os.environ["o365-passwd"],
    'url': "https://scouterna.sharepoint.com" + "/sites/StabenJam21"
}
file_id = '5A578FC8-26D7-45B9-A0BD-D5BA76A34575'

def run(file_path):
    download_path = file_path
    user_credentials = UserCredential(settings['username'], settings['password'])

    ctx = ClientContext(settings['url']).with_credentials(user_credentials)
    target_web = ctx.web.get().execute_query()
    print(target_web.url)

    with open(download_path, "wb") as local_file:
        ctx.web.get_file_by_id(file_id).download(local_file).execute_query()
    print("[Ok] file has been downloaded: {0}".format(download_path))