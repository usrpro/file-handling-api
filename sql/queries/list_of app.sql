select fs.id, fs.name, fs.remote_addr, fs.application, fs.created 
from files_stored fs 
where fs.application = $1 order by id;