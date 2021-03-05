from shutil import which
import psycopg2
from psycopg2.errors import UndefinedTable


from behave import given, then, when


@given(u"the system is in default state")
def system_in_default_state(context):
    """Check the default system state."""
    pass


@when(u"I look for executable file {filename}")
def look_for_executable_file(context, filename):
    """Try to find given executable file on PATH."""
    context.filename = filename
    context.found = which(filename)


@then(u"I should find that file on PATH")
def file_was_found(context):
    """Check if the file was found on PATH."""
    assert context.found is not None, \
        "executable filaname '{}' is not on PATH".format(context.filename)


@when(u"I connect to database named {database} as user {user} with password {password}")
def connect_to_database(context, database, user, password):
    """Perform connection to selected database."""
    connection_string = "dbname={} user={} password={}".format(database, user, password)
    context.connection = psycopg2.connect(connection_string)


@then(u"I should be able to connect to such database")
def check_connection(context):
    """Chck the connection to database."""
    assert context.connection is not None, "connection should be established"


@when(u"I close database connection")
def disconnect_from_database(context):
    """Close the connection to database."""
    context.connection.close()
    context.connection = None


@then(u"I should be disconnected")
def check_disconnection(context):
    """Check that the connection has been closed."""
    assert context.connection is None, "connection should be closed"


@when(u"I look for the table {table} in database")
def look_for_table(context, table):
    """Try to find a table in database."""
    cursor = context.connection.cursor()
    try:
        cursor.execute("SELECT 1 from {}".format(table))
        v = cursor.fetchone()
        context.table_found = True
    except UndefinedTable as e:
        context.table_found = False


@then(u"I should not be able to find it")
def check_table_existence(context):
    """Check the table existence in the database."""
    assert context.table_found is False, "table should not exist"
