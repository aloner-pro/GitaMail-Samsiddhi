import yaml
from pyairtable import Api

with open('test.yaml', 'r') as file:
    config = yaml.safe_load(file)

api = Api(config['airtabletoken'])
table = api.table(config['baseid'], config['tableid'])
data = table.all()

def get_mail(data):
    mail_list = []
    for i in data:
        if i['fields']['Interested'] == 'Yes':
            mail_list.append(i['fields']['Email'])

    return mail_list

print(get_mail(data))
