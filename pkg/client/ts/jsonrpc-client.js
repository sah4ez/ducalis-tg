
export class JSONRPCError extends Error {
	constructor(message, name, code, data) {
	  	super(message);
	  	this.name = name;
	  	this.code = code;
		this.data = data;
	}
}

class JSONRPCScheduler {
	/**
	 *
	 * @param {*} transport
	 */
	constructor(transport) {
	  this._transport = transport;
	  this._requestID = 0;
	  this._scheduleRequests = {};
	  this._commitTimerID = null;
	  this._beforeRequest = null;
	}
	beforeRequest(fn) {
	  this._beforeRequest = fn;
	} 
	__scheduleCommit() {
	  if (this._commitTimerID) {
		clearTimeout(this._commitTimerID);
	  }
	  this._commitTimerID = setTimeout(() => {
		this._commitTimerID = null;
		const scheduleRequests = { ...this._scheduleRequests };
		this._scheduleRequests = {};
		let requests = [];
		for (let key in scheduleRequests) {
		  requests.push(scheduleRequests[key].request);
		}
		this.__doRequest(requests)
		  .then((responses) => {
			for (let i = 0; i < responses.length; i++) {
              const schedule = scheduleRequests[responses[i].id];
			  if (responses[i].error) {
				schedule.reject(responses[i].error);
				continue;
			  }
			  schedule.resolve(responses[i].result);
			}
		  })
         .catch((e) => {
           for (let key in requests) {
             if (!requests.hasOwnProperty(key)) {
               continue;
             }
             if (scheduleRequests.hasOwnProperty(requests[key].id)) {
               scheduleRequests[requests[key].id].reject(e)
             }
           }
         });
	  }, 0);
	}
	makeJSONRPCRequest(id, method, params) {
	  return {
		jsonrpc: "2.0",
		id: id,
		method: method,
		params: params,
	  };
	}
	/**
    * @param {string} method
    * @param {Object} params
    * @returns {Promise<*>}
    */
	__scheduleRequest(method, params) {
	  const p = new Promise((resolve, reject) => {
		const request = this.makeJSONRPCRequest(
		  this.__requestIDGenerate(),
		  method,
		  params
		);
		this._scheduleRequests[request.id] = {
		  request,
		  resolve,
		  reject,
		};
	  });
	  this.__scheduleCommit();
	  return p;
	}
	__doRequest(request) {
	  return this._transport.doRequest(request);
	}
	__requestIDGenerate() {
	  return ++this._requestID;
	}
 }
class JSONRPCClientAuthService {
constructor(transport) {
this.scheduler = new JSONRPCScheduler(transport);
}

/**
* Register
*
* @param {RegisterRequest} Req
* @return {PromiseLike<{User: User,Token: string}>}
**/
register(req) {
return this.scheduler.__scheduleRequest("authService.register", {req:req}).catch(e => { throw authServiceregisterConvertError(e); })
}
/**
* Login
*
* @param {string} Email
* @param {string} Password
* @return {PromiseLike<{User: User,Token: string}>}
**/
login(email,password) {
return this.scheduler.__scheduleRequest("authService.login", {email:email,password:password}).catch(e => { throw authServiceloginConvertError(e); })
}
/**
* OAuth
*
* @param {string} Provider
* @param {string} RedirectURL
* @return {PromiseLike<{AuthURL: string}>}
**/
oAuthAuthorize(provider,redirectURL) {
return this.scheduler.__scheduleRequest("authService.oAuthAuthorize", {provider:provider,redirectURL:redirectURL}).catch(e => { throw authServiceoAuthAuthorizeConvertError(e); })
}
/**
* OAuth
*
* @param {string} Provider
* @param {string} Code
* @param {string} State
* @return {PromiseLike<{User: User,Token: string}>}
**/
oAuthCallback(provider,code,state) {
return this.scheduler.__scheduleRequest("authService.oAuthCallback", {provider:provider,code:code,state:state}).catch(e => { throw authServiceoAuthCallbackConvertError(e); })
}
/**
* Get
*
* @param {string} UserID
* @return {PromiseLike<{User: User}>}
**/
getMe(userID) {
return this.scheduler.__scheduleRequest("authService.getMe", {userID:userID}).catch(e => { throw authServicegetMeConvertError(e); })
}
/**
* Refresh
*
* @param {string} RefreshToken
* @return {PromiseLike<{Token: string}>}
**/
refreshToken(refreshToken) {
return this.scheduler.__scheduleRequest("authService.refreshToken", {refreshToken:refreshToken}).catch(e => { throw authServicerefreshTokenConvertError(e); })
}
}

class JSONRPCClientIntegrationService {
constructor(transport) {
this.scheduler = new JSONRPCScheduler(transport);
}

/**
* Create
*
* @param {CreateGitHubIntegrationRequest} Req
* @return {PromiseLike<{Integration: Integration}>}
**/
createGitHub(req) {
return this.scheduler.__scheduleRequest("integrationService.createGitHub", {req:req}).catch(e => { throw integrationServicecreateGitHubConvertError(e); })
}
/**
* Create
*
* @param {CreateJiraIntegrationRequest} Req
* @return {PromiseLike<{Integration: Integration}>}
**/
createJira(req) {
return this.scheduler.__scheduleRequest("integrationService.createJira", {req:req}).catch(e => { throw integrationServicecreateJiraConvertError(e); })
}
/**
* Create
*
* @param {CreateLinearIntegrationRequest} Req
* @return {PromiseLike<{Integration: Integration}>}
**/
createLinear(req) {
return this.scheduler.__scheduleRequest("integrationService.createLinear", {req:req}).catch(e => { throw integrationServicecreateLinearConvertError(e); })
}
/**
* List
*
* @param {string} WorkspaceID
* @return {PromiseLike<{Integrations: Array<Integration>}>}
**/
list(workspaceID) {
return this.scheduler.__scheduleRequest("integrationService.list", {workspaceID:workspaceID}).catch(e => { throw integrationServicelistConvertError(e); })
}
/**
* Delete
*
* @param {string} Id
**/
delete(id) {
return this.scheduler.__scheduleRequest("integrationService.delete", {id:id}).catch(e => { throw integrationServicedeleteConvertError(e); })
}
/**
* Sync
*
* @param {string} Id
* @return {PromiseLike<{Result: SyncResult}>}
**/
sync(id) {
return this.scheduler.__scheduleRequest("integrationService.sync", {id:id}).catch(e => { throw integrationServicesyncConvertError(e); })
}
/**
* Get
*
* @param {string} Id
* @return {PromiseLike<{Status: SyncStatus}>}
**/
getSyncStatus(id) {
return this.scheduler.__scheduleRequest("integrationService.getSyncStatus", {id:id}).catch(e => { throw integrationServicegetSyncStatusConvertError(e); })
}
}

class JSONRPCClientTaskService {
constructor(transport) {
this.scheduler = new JSONRPCScheduler(transport);
}

/**
* Create
*
* @param {CreateTaskRequest} Req
* @return {PromiseLike<{Task: Task}>}
**/
create(req) {
return this.scheduler.__scheduleRequest("taskService.create", {req:req}).catch(e => { throw taskServicecreateConvertError(e); })
}
/**
* Get
*
* @param {string} Id
* @return {PromiseLike<{Task: Task}>}
**/
get(id) {
return this.scheduler.__scheduleRequest("taskService.get", {id:id}).catch(e => { throw taskServicegetConvertError(e); })
}
/**
* Update
*
* @param {string} Id
* @param {UpdateTaskRequest} Req
* @return {PromiseLike<{Task: Task}>}
**/
update(id,req) {
return this.scheduler.__scheduleRequest("taskService.update", {id:id,req:req}).catch(e => { throw taskServiceupdateConvertError(e); })
}
/**
* Delete
*
* @param {string} Id
**/
delete(id) {
return this.scheduler.__scheduleRequest("taskService.delete", {id:id}).catch(e => { throw taskServicedeleteConvertError(e); })
}
/**
* List
*
* @param {ListTasksRequest} Req
* @return {PromiseLike<{Tasks: Array<Task>,Total: number}>}
**/
list(req) {
return this.scheduler.__scheduleRequest("taskService.list", {req:req}).catch(e => { throw taskServicelistConvertError(e); })
}
/**
* Set
*
* @param {string} TaskID
* @param {Object<string,number>} Scores
* @return {PromiseLike<{Task: Task}>}
**/
setScores(taskID,scores) {
return this.scheduler.__scheduleRequest("taskService.setScores", {taskID:taskID,scores:scores}).catch(e => { throw taskServicesetScoresConvertError(e); })
}
/**
* Vote
*
* @param {string} TaskID
* @param {string} UserID
* @param {number} Weight
* @return {PromiseLike<{Task: Task}>}
**/
vote(taskID,userID,weight) {
return this.scheduler.__scheduleRequest("taskService.vote", {taskID:taskID,userID:userID,weight:weight}).catch(e => { throw taskServicevoteConvertError(e); })
}
/**
* Remove
*
* @param {string} TaskID
* @param {string} UserID
* @return {PromiseLike<{Task: Task}>}
**/
removeVote(taskID,userID) {
return this.scheduler.__scheduleRequest("taskService.removeVote", {taskID:taskID,userID:userID}).catch(e => { throw taskServiceremoveVoteConvertError(e); })
}
/**
* Estimate
*
* @param {string} TaskID
* @param {string} UserID
* @param {number} Value
* @return {PromiseLike<{Task: Task}>}
**/
estimate(taskID,userID,value) {
return this.scheduler.__scheduleRequest("taskService.estimate", {taskID:taskID,userID:userID,value:value}).catch(e => { throw taskServiceestimateConvertError(e); })
}
/**
* Set
*
* @param {string} TaskID
* @param {Array<string>} DependencyIDs
* @return {PromiseLike<{Task: Task}>}
**/
setDependencies(taskID,dependencyIDs) {
return this.scheduler.__scheduleRequest("taskService.setDependencies", {taskID:taskID,dependencyIDs:dependencyIDs}).catch(e => { throw taskServicesetDependenciesConvertError(e); })
}
/**
* Get
*
* @param {string} WorkspaceID
* @param {number} Limit
* @param {number} Offset
* @return {PromiseLike<{Tasks: Array<TaskWithRank>}>}
**/
getRanked(workspaceID,limit,offset) {
return this.scheduler.__scheduleRequest("taskService.getRanked", {workspaceID:workspaceID,limit:limit,offset:offset}).catch(e => { throw taskServicegetRankedConvertError(e); })
}
}

class JSONRPCClientWorkspaceService {
constructor(transport) {
this.scheduler = new JSONRPCScheduler(transport);
}

/**
* Create
*
* @param {CreateWorkspaceRequest} Req
* @return {PromiseLike<{Workspace: Workspace}>}
**/
create(req) {
return this.scheduler.__scheduleRequest("workspaceService.create", {req:req}).catch(e => { throw workspaceServicecreateConvertError(e); })
}
/**
* Get
*
* @param {string} Id
* @return {PromiseLike<{Workspace: Workspace}>}
**/
get(id) {
return this.scheduler.__scheduleRequest("workspaceService.get", {id:id}).catch(e => { throw workspaceServicegetConvertError(e); })
}
/**
* Update
*
* @param {string} Id
* @param {UpdateWorkspaceRequest} Req
* @return {PromiseLike<{Workspace: Workspace}>}
**/
update(id,req) {
return this.scheduler.__scheduleRequest("workspaceService.update", {id:id,req:req}).catch(e => { throw workspaceServiceupdateConvertError(e); })
}
/**
* Delete
*
* @param {string} Id
**/
delete(id) {
return this.scheduler.__scheduleRequest("workspaceService.delete", {id:id}).catch(e => { throw workspaceServicedeleteConvertError(e); })
}
/**
* List
*
* @param {string} UserID
* @param {number} Limit
* @param {number} Offset
* @return {PromiseLike<{Workspaces: Array<Workspace>,Total: number}>}
**/
list(userID,limit,offset) {
return this.scheduler.__scheduleRequest("workspaceService.list", {userID:userID,limit:limit,offset:offset}).catch(e => { throw workspaceServicelistConvertError(e); })
}
/**
* Set
*
* @param {string} WorkspaceID
* @param {ScoringConfig} Config
* @return {PromiseLike<{Workspace: Workspace}>}
**/
setScoringConfig(workspaceID,config) {
return this.scheduler.__scheduleRequest("workspaceService.setScoringConfig", {workspaceID:workspaceID,config:config}).catch(e => { throw workspaceServicesetScoringConfigConvertError(e); })
}
/**
* Invite
*
* @param {string} WorkspaceID
* @param {string} Email
* @param {string} Role
* @return {PromiseLike<{Member: Member}>}
**/
inviteMember(workspaceID,email,role) {
return this.scheduler.__scheduleRequest("workspaceService.inviteMember", {workspaceID:workspaceID,email:email,role:role}).catch(e => { throw workspaceServiceinviteMemberConvertError(e); })
}
/**
* List
*
* @param {string} WorkspaceID
* @return {PromiseLike<{Members: Array<Member>}>}
**/
listMembers(workspaceID) {
return this.scheduler.__scheduleRequest("workspaceService.listMembers", {workspaceID:workspaceID}).catch(e => { throw workspaceServicelistMembersConvertError(e); })
}
/**
* Remove
*
* @param {string} WorkspaceID
* @param {string} MemberID
**/
removeMember(workspaceID,memberID) {
return this.scheduler.__scheduleRequest("workspaceService.removeMember", {workspaceID:workspaceID,memberID:memberID}).catch(e => { throw workspaceServiceremoveMemberConvertError(e); })
}
}

class JSONRPCClient {
constructor(transport) {
this.authService = new JSONRPCClientAuthService(transport);
this.integrationService = new JSONRPCClientIntegrationService(transport);
this.taskService = new JSONRPCClientTaskService(transport);
this.workspaceService = new JSONRPCClientWorkspaceService(transport);
}
}
export default JSONRPCClient

function authServiceregisterConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function authServiceloginConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function authServiceoAuthAuthorizeConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function authServiceoAuthCallbackConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function authServicegetMeConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function authServicerefreshTokenConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function integrationServicecreateGitHubConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function integrationServicecreateJiraConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function integrationServicecreateLinearConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function integrationServicelistConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function integrationServicedeleteConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function integrationServicesyncConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function integrationServicegetSyncStatusConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServicecreateConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServicegetConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServiceupdateConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServicedeleteConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServicelistConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServicesetScoresConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServicevoteConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServiceremoveVoteConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServiceestimateConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServicesetDependenciesConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function taskServicegetRankedConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function workspaceServicecreateConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function workspaceServicegetConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function workspaceServiceupdateConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function workspaceServicedeleteConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function workspaceServicelistConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function workspaceServicesetScoringConfigConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function workspaceServiceinviteMemberConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function workspaceServicelistMembersConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
function workspaceServiceremoveMemberConvertError(e) {
switch(e.code) {
default:
return new JSONRPCError(e.message, "UnknownError", e.code, e.data);
}
}
/**
* @typedef {Object} ScoringConfig
* @property {string} type
* @property {Array<Criterion>} criteria
* @property {string} formula
*/

/**
* @typedef {Object} Workspace
* @property {string} updatedAt
* @property {string} id
* @property {string} name
* @property {string} description
* @property {string} ownerId
* @property {Object<ScoringConfig>} scoring
* @property {string} createdAt
*/

/**
* @typedef {Object} Member
* @property {string} name
* @property {string} role
* @property {number} voteWeight
* @property {string} joinedAt
* @property {string} id
* @property {string} workspaceId
* @property {string} userId
* @property {string} email
*/

/**
* @typedef {Object} CreateGitHubIntegrationRequest
* @property {string} repo
* @property {boolean} autoSync
* @property {boolean} syncLabels
* @property {string} labelFilter
* @property {string} workspaceId
* @property {string} name
* @property {string} token
* @property {string} owner
*/

/**
* @typedef {Object} Integration
* @property {string} workspaceId
* @property {string} type
* @property {string} name
* @property {Object<string, interface>} config
* @property {string} syncStatus
* @property {string} syncError
* @property {number} syncInterval
* @property {string} lastSyncAt
* @property {boolean} autoSync
* @property {string} createdAt
* @property {string} updatedAt
* @property {string} id
*/

/**
* @typedef {Object} CreateTaskRequest
* @property {string} externalId
* @property {string} externalUrl
* @property {string} title
* @property {string} description
* @property {Object<string, number>} scores
* @property {Array<string>} dependencies
* @property {string} priority
* @property {Array<string>} labels
* @property {string} workspaceId
* @property {string} externalType
* @property {string} status
* @property {string} assigneeId
* @property {Object<string, interface>} metadata
*/

/**
* @typedef {Object} Vote
* @property {string} userId
* @property {number} weight
* @property {string} createdAt
*/

/**
* @typedef {Object} TaskWithRank
* @property {Object<Task>} Task
* @property {number} rank
* @property {number} percentile
*/

/**
* @typedef {Object} Criterion
* @property {Array<string>} options
* @property {number} weight
* @property {boolean} required
* @property {string} id
* @property {string} name
* @property {string} description
* @property {string} type
* @property {Object<Scale>} scale
*/

/**
* @typedef {Object} CreateWorkspaceRequest
* @property {string} name
* @property {string} description
* @property {Object<ScoringConfig>} scoring
*/

/**
* @typedef {Object} SyncResult
* @property {number} tasksDeleted
* @property {Array<SyncError>} errors
* @property {number} durationMs
* @property {string} nextScheduledAt
* @property {string} integrationId
* @property {string} status
* @property {number} tasksCreated
* @property {number} tasksUpdated
*/

/**
* @typedef {Object} UpdateTaskRequest
* @property {string} title
* @property {string} description
* @property {string} status
* @property {string} priority
* @property {Array<string>} labels
* @property {string} assigneeId
* @property {Object<string, interface>} metadata
*/

/**
* @typedef {Object} UpdateWorkspaceRequest
* @property {string} name
* @property {string} description
* @property {Object<ScoringConfig>} scoring
*/

/**
* @typedef {Object} Estimation
* @property {string} userId
* @property {number} value
* @property {string} unit
* @property {string} createdAt
*/

/**
* @typedef {Object} Task
* @property {Array<Estimation>} estimations
* @property {Array<string>} dependencies
* @property {string} priority
* @property {Array<string>} labels
* @property {string} assigneeId
* @property {string} createdBy
* @property {Object<string, interface>} metadata
* @property {string} externalType
* @property {string} externalUrl
* @property {Object<string, number>} scores
* @property {Array<Vote>} votes
* @property {string} createdAt
* @property {string} updatedAt
* @property {string} title
* @property {string} description
* @property {number} finalScore
* @property {string} id
* @property {string} externalId
* @property {string} status
* @property {string} workspaceId
*/

/**
* @typedef {Object} User
* @property {string} id
* @property {string} email
* @property {string} name
* @property {string} avatarUrl
* @property {string} createdAt
*/

/**
* @typedef {Object} CreateJiraIntegrationRequest
* @property {string} username
* @property {string} apiToken
* @property {string} projectKey
* @property {string} jqlFilter
* @property {boolean} autoSync
* @property {string} workspaceId
* @property {string} name
* @property {string} baseUrl
*/

/**
* @typedef {Object} CreateLinearIntegrationRequest
* @property {string} apiKey
* @property {string} teamId
* @property {boolean} autoSync
* @property {string} workspaceId
* @property {string} name
*/

/**
* @typedef {Object} SyncError
* @property {string} externalId
* @property {string} error
*/

/**
* @typedef {Object} ListTasksRequest
* @property {string} workspaceId
* @property {string} assigneeId
* @property {boolean} hasVotes
* @property {number} limit
* @property {number} offset
* @property {string} status
* @property {Array<string>} labels
* @property {boolean} hasScore
* @property {number} minScore
* @property {number} maxScore
* @property {string} search
* @property {string} sortBy
* @property {boolean} sortDesc
*/

/**
* @typedef {Object} Scale
* @property {number} min
* @property {number} max
*/

/**
* @typedef {Object} RegisterRequest
* @property {string} name
* @property {string} email
* @property {string} password
*/

/**
* @typedef {Object} SyncStatus
* @property {string} lastError
* @property {string} nextSyncAt
* @property {number} totalTasks
* @property {number} lastDurationMs
* @property {string} integrationId
* @property {string} status
* @property {string} lastSyncAt
* @property {string} lastSyncResult
*/

