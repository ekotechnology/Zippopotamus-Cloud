from bottle import route, default_app, static_file, view, response, run, request
from pymongo import Connection
from pymongo.database import Database
from bson.son import SON
import json
import os
import redis
import pickle

os.chdir(os.path.dirname(__file__))

r = redis.StrictRedis(host='localhost', port=6379, db=0)

class static_files():

    '''
    Serve Static Files for homepage. This is done with by
    mapping all the /static/* folder to be served
    '''
    @route('/static/:path#.+#')
    def serve(path):
        return static_file(path, root='./static')

    '''
    Serve cross domain on seperate route
    '''
    @route('/crossdomain.xml')
    def xdomain():
        return static_file('crossdomain.xml', root='./static')


class index():
    '''
    Class for dynamic URLs and queries
    '''

    @route('/')
    @view('index')
    def homepage():
        return {}

    @route('/:country/:post', method='GET')
    def index(country, post):
        (isFound, results) = standard_query(country, post)
        configure(response)
        if (isFound is False):    # If standard query got >= 1 hit
            response.status = 404      # Set Status to a 404
        return results               # Return empty JSON String

    @route('/nearby/:country/:post', method='GET')
    def find_nearby(country, post):
        (isFound, result) = nearby_zip(country, post)
        configure(response)
        if (isFound is False):
            response.status = 404
        return result

    @route('/:country/:state/:place', method='GET')
    def find_postcode(country, state, place):
        (isFound, results) = place_query(country, state, place)
        configure(response)
        if (isFound is False):
            response.status = 404
        return results


def configure(response):
    '''
    Configure:
        Given a response dict, configure will
        set the appropriate headers
    '''
    response['Content-Type'] = 'application/json'        # Specify MIME type to be JSON
    response['charset'] = 'UTF-8'                        # Speciify Charset for browser viewing
    response['Access-Control-Allow-Origin'] = '*'        # Enables CORS for XHR request
    response['Vary'] = 'Accept-Encoding'
    pass


def nearby_zip(country, code):
    '''
    Looks up nearby postcodes given country and postal code
    Returns results for JSON response
    '''
    cKey = 'nearby.' + country.upper() + '.' + code

    if (r.get(cKey)):
        stat_count(True, True)
        return (True, r.get(cKey))
    else:
        post = list(db['nearby'].find({'post code': code.upper(),
                                       'country abbreviation': country.upper()}))
    if len(post) >= 1:
        lon = float(post[0]['longitude'])
        lat = float(post[0]['latitude'])
        (success, nearby) = nearby_query(lat, lon)

        if success:
            response = {'near latitude': lat,      # Record the query lat
                        'near longitude': lon,      # Record the query lat
                        'nearby': nearby[1:]}
            content = json.dumps(response)
            stat_count(True, False)
            r.set(cKey, content)
            return (True, content)

    content = json.dumps({})
    stat_count(False)
    return (False, content)


def nearby_query(lat, lon, num = 11):
    '''
    GeoSpatial  Query,
    Given a specific latitude or longitude, this returns the closes 11 zipcodes
    '''
    cKey = 'nearby.' + str(lat) + '.' + str(lon)

    if r.get(cKey):
            stat_count(True, True)
            return (True, pickle.loads(r.get(cKey)))
    else:
        nearby = db.command(SON([('geoNear', 'nearby'),           # Geospatial search
                                ('near', [lon, lat]),            # near given coordinates
                                ('distanceMultiplier', 3959),   # Return values in miles
                                ('spherical', True),              # Spherical
                                ('num', num)]))                  # Results to return

    if nearby['ok'] > 0:
        results = list()
        for records in nearby['results']:
            places = records['obj']
            places['distance'] = records['dis']
            del places['loc']                       # Remove Coordinate info
            try:
                del places['_id']                       # Remove mongo_id
            except:
                pass
            del places['latitude']                  # Remove string latitude
            del places['longitude']                 # Remove string long
            del places['country']                   # Remove country
            del places['country abbreviation']      # Remove abbrevation
            results.append(places)
            r.set(cKey, pickle.dumps(results))
            stat_count(True, False)
        return (True, results)
    else:
        stat_count(False)
        return (False, "")


def standard_query(country, code):
    '''
    Standard_query returns a JSON data if there are matching places, for a
    given country abbreviation and zip code
    '''
    cKey = 'standard.' + country.upper() + '.' + code

    if r.get(cKey):
        stat_count(True, True)
        return (True, r.get(cKey))
    else:
        result = list(db['global'].find({'country abbreviation': country.upper(),
                                         'post code': code.upper()}))

    if len(result) < 1:
        content = json.dumps({})        # return empty json string
        stat_count(False)
        return (False, content)
    else:
        country_name = result[0]['country']                    # Capture country
        country_abbv = result[0]['country abbreviation']       # Country abbrev.
        post_code = result[0]['post code']                # and post code
        isFound = True                                      # if Match found

        for places in result:                                # Remove from each result
            try:
                del places['_id']                                # mongo id
            except:
                pass
            del places['post code']                          # post code
            del places['country']                            # country
            del places['country abbreviation']               # country abbrev. information

        content = json.dumps({'country': country_name,                  # Return unique fields
                              'country abbreviation': country_abbv,
                              'post code': post_code,
                              'places': result})                 # Using pymongo json settings
        r.set(cKey, content)
        stat_count(True, False)
        return (isFound, content)    # Return True and JSON results


def place_query(country, state, place):
    '''
    Place_query returns JSON data if there are matching postcodes for a given
    country abbreviation, state abbreviation, and place/city
    '''
    cKey = "place." + country.upper() + '.' + state.upper() + '.' + place.upper()

    if (r.get(cKey)):
        stat_count(True, True)
        return (True, r.get(cKey))
    else:
        result = list(db['global'].find({'country abbreviation': country.upper(),
                                         'state abbreviation': state.upper(),
                                         'place name': {'$regex': place, '$options': '-i'}
                                         }))

    if len(result) < 1:
        content = json.dumps({})   # Empty JSON string
        stat_count(False)
        return (False, content)
    else:
        country_name = result[0]['country']
        country_abbv = result[0]['country abbreviation']
        state = result[0]['state']
        state_abbv = result[0]['state abbreviation']
        place = result[0]['place name']
        isFound = True
        for places in result:                           # Remove from each result
            try:
                del places['_id']                           # Mongo ID
            except:
                pass
            del places['state']                         # State
            del places['state abbreviation']            # State abbreviation
            del places['country']                       # Country
            del places['country abbreviation']          # Country abbreviation

        content = json.dumps({
            'country': country_name,
            'country abbreviation': country_abbv,
            'state': state,
            'state abbreviation': state_abbv,
            'place name': place,
            'places': result})
        stat_count(True, False)
        r.set(cKey, content)
        return (isFound, content)   # Return True and JSON results


def stat_count(found, cacheHit=None):
    '''
    Add to Redis request counter, overall request numbers
    as well as unique IP requests, and AJAX vs. other
    '''
    r.incr('request.count')
    r.hincrby('request.hosts', request.remote_addr.replace('.', '-'), 1)

    if (request.is_xhr):
        r.incr('request.xhr')

    if (not found):
        r.incr('request.notFound')
    else:
        if (cacheHit):
            r.incr('request.cacheHit')
            response['X-CACHE'] = 'hit'
        else:
            r.incr('request.cacheMiss')
            response['X-CACHE'] = 'miss'

# PRESENT GLOBAL
ZIP = 'zip'
NEARBY = 'nearby'

connection = Connection()
db = Database(connection, 'zip')                             # Get handle to ZIP db

application = default_app()                                 # WSGI application
#run (host='localhost', port=8080)                          # Local Testing
