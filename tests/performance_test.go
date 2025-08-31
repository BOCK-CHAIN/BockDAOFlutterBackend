package tests

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/core"
	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/dao"
	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PerformanceTestSuite contains performance benchmarks and load tests
type PerformanceTestSuite struct {
	daoInstance *dao.DAO
	blockchain  *core.Blockchain
	logger      log.Logger
	cleanup     func()
}

// NewPerformanceTestSuite creates a new performance test suite
func NewPerformanceTestSuite(t *testing.T) *PerformanceTestSuite {
	logger := log.NewNopLogger()

	genesis := createTestGenesisBlock(t)
	bc, err := core.NewBlockchain(logger, genesis)
	require.NoError(t, err)

	daoInstance := dao.NewDAO("PERF", "Performance Test Token", 18)

	// Initialize with substantial test distribution
	testDistribution := map[string]uint64{
		"performance_treasury": 100000000, // 100M tokens
	}
	err = daoInstance.InitialTokenDistribution(testDistribution)
	require.NoError(t, err)

	return &PerformanceTestSuite{
		daoInstance: daoInstance,
		blockchain:  bc,
		logger:      logger,
		cleanup: func() {
			// Cleanup resources
		},
	}
}

// TestHighThroughputOperations tests system performance under high load
func TestHighThroughputOperations(t *testing.T) {
	suite := NewPerformanceTestSuite(t)
	defer suite.cleanup()

	t.Run("ProposalCreationThroughput", suite.testProposalCreationThroughput)
	t.Run("VotingThroughput", suite.testVotingThroughput)
	t.Run("TokenTransferThroughput", suite.testTokenTransferThroughput)
	t.Run("DelegationThroughput", suite.testDelegationThroughput)
	t.Run("TreasuryOperationThroughput", suite.testTreasuryOperationThroughput)
	t.Run("ConcurrentOperationMix", suite.testConcurrentOperationMix)
}

// TestScalabilityLimits tests system behavior at scale limits
func TestScalabilityLimits(t *testing.T) {
	suite := NewPerformanceTestSuite(t)
	defer suite.cleanup()

	t.Run("LargeUserBase", suite.testLargeUserBase)
	t.Run("MassiveProposalVolume", suite.testMassiveProposalVolume)
	t.Run("ComplexDelegationChains", suite.testComplexDelegationChains)
	t.Run("HighFrequencyVoting", suite.testHighFrequencyVoting)
	t.Run("MemoryUsageUnderLoad", suite.testMemoryUsageUnderLoad)
}

// TestResourceUtilization tests resource usage patterns
func TestResourceUtilization(t *testing.T) {
	suite := NewPerformanceTestSuite(t)
	defer suite.cleanup()

	t.Run("CPUUtilization", suite.testCPUUtilization)
	t.Run("MemoryEfficiency", suite.testMemoryEfficiency)
	t.Run("GarbageCollectionImpact", suite.testGarbageC
ecks)
}

// Performance t

func (suite *PerformanceTestSuite) testg.T) {
	// Setup creators
	numCreators := 10
	creators := make([]crypto.PrivateKey, nuors)
	frs {
	teKey()
	}
	s...)

	// Benchmark proposal creation
	numProposals := 1000
	start := time.Now()

	for i := 0; i < numProposals; i++ {
		creator := creators[i%numCreators]

		proposalTx := &dao.ProposalTx{
			Fee:          200,
			Title:        fmt.S
			Description:  fmt.Sprintf("
			Peneral,
			VotingType:   dao.VotingT
		
			EndTime:      time.Now
	   1000,
			MetadataHash: suite.randomHash(),
		}

	
		err := suite.daoInstance.ProcessDAOTransactioosalHash)
	t, err)
	}

	duration := time.Since(start)
	throughput := float64(numProposds()

	t.Logf("Proposal Creation Throughput: %.2f proposals/ut)
	

	// Performions
	assert.Greater(t, throughput, 50.0, "Should achieve c")
	assert.Less(t, du")

	// Verify all proposals were 
	proposals := suite.daols()
	assert.GreaterOrEqual(t, len
}

funcng.T) {
	//
	creator := crypto.GeneratePriv
	numVoters := 500
	voters := make([]cryrs)
	for i := range voters {
		voters[i] = crypto.Geney()
	}
	suite.setupTestUsers(t, append(voters, creator)...)

	// Create proposal
	prox{
		Fee:          200,
		Title:        "Voti
		Description
		
	
		StartTime:    t00,
		EndTime:  
		Threshold1000,
		MetadataHash: ssh(),
	}

	proposalHash := suite.generater)
	err := suite.daoInstance.ProcessDAash)
	r err)

	// Benchmark voting
	start := time.Now()

	for i, voter := range {
		choice := dao.VoteChoics
		if i%3 == 0 {
			eNo
		

		voteTx := &dao.VoteTx{
			Fee:        100,
			ProposalID: proposalHash,
		
	100,
			Reason:     fmt.Sprintf("Perfor
	

		voteHash := suite.generateTxHash(vot
		err := suite.daoInstanHash)
		require.NoError(t, err)
	}

	duration := time.Since(start)
	ds()

	ghput)
ters))

	// Performance assertions
	assert.Greater(")

	// Verify all votes were recded
	votes, err := suite.daoInstance.GetVotes(plHash)
	require.NoError(t, err)
	assert.Len(t, votes, numVoters)
}

func (suite *Performance
	rs
	numUsers := 100
	users := make([]Users)
	for i := range users {
		users[i] = crypto.Ge)
	}
	suite.setupTestUsers..)

	/ers
	numTransfers := 1000
	start := time.Now()

	
		from := users[i%numUsers]
	]

		err := suite.daoInsta10)
	err)
	}

	duration := time.Since(start)
	throughput := float64(numTransfers

	t)
	t.Logf("Average time per transfers))

	// Performance asstions
	assert.Greater(t, throughput, 200.0,)
}

func (suite *PerformanceTestng.T) {
	//s
	numUsers := 200
	users := make([]crypto.PrivateKey, numUsers)
	f
	
	}
	suite.setupTestUsers(t, users...)

	// Benchmark delegations
	numDelegations := numUse
	stw()

	for i := 0; i < numDelegatio
		delegator := users[i*2]
		delegate := users[i*2+1]

		x{
	   200,
			Delegate: delegate.PublicKey(),
	
			Revoke:   false,
		}

	r)
		err := suite.daoInstance.ProcessDAOTransaction(delegationTx, delegator.PublicKey(), delegationHash)
	r)


	duration := time.Since(start)
	throughput := f)

	t.Logf("Delegation Throughput: %.2f delegations/sec",put)
	t.Logf("Average time per delegation: %vs))

	// Performance assertions
	a
}

f {
	// Setup treasury
	
	signers := make([]crypto.PrivateKeyigners)
	signerPubKeys := make([]crypto.Publicers)
	rs {
		signers[i] = crypto.GeneratePrivateKey()
		signerPubKeys[i] = signe()
	}

	e
	require.NoError(t, err)

	suite.daoInstanceokens

	// Benchmark treasations
	numOperations := 50
	reci
	for i := range recipients {
		reKey()
	}

	sw()

	for i := 0; i < numOperations; i+
		treasuryTx := ryTx{
			Fee:          500,
			Recipient:    recipy(),
			Amount:       10000,
			Purpose:      fmt.Sprint
			{},
		 3,
		}

		txHash := suite.generateT0])
		err := suite.daoInstanc txHash)
		

		// Sign with required signers
		for j := 0; j < 3; j++ {
	ers[j])
			require.NoError(t, err)
		}

		// Execute
		on(txHash)
		require.NoError(t, err)
	}

	d)
	

	
	t.Logf("Average time per

	ertions
	assert.Greater(t, throughput, 10.0, "Should achieve at least 10 treasury operations/sec")
}

func (suite *PerformanceTestSuite) testConcurrentOperag.T) {
	// Setup users
	numUsers := 50
	
	for i := range users {
		users[i] = crypto.GeneratePrivateKey()
	}
	suite.setupTestUsers(t, users...)

	
	signers := users[:5]
	signerPubKeys := make([]cry
	for i, signer := range signers {
		ublicKey()
	
	err := suite.daoInstance.InitializeTreasury(signerPubKey
	(t, err)
	suite.daoInstance.AddT0)

	// Concurrent mixed operations
	var wg sync.WaitGroup
	start := time.Now()

	/orkers
	wg.Add(1)
	
		defer wg.Done()
	
			creator := users[i%numUsers]
			proposalTx := &dao.ProposalTx{
	
				Title:        fmt.Spri", i),
				Description:  "Mixed operation test proposal",
		eral,
				VotingType:   dao.VotingTypeSimple,
				StartTime:    time.Now().Unix() - 100,
				EndTime:      time.Now().Unix(),
		
				MetadataHash: suite.randomHash(),
			}

			proposalHash := suite.generateTxHash(
			sh)
		}
	}()

	// Token transfer workers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			from := users[i%nUsers]
			tomUsers]
			suite.daoInstance.TransferTok, 5)
		}
	}()

	/ers
	.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 25; i++ {
			delegator := users[ers]
			delegate := users[(i*2+1)%]
			delegationTx := &daash
}	return h 256)
	}
byte(i %h[i] = h {
		hasnge hasfor i := ra2]byte{}
		hash := [3pes.Hash {
) tydomHash(rante) ceTestSui*Performansuite  (

funcash
}rn htu	re
:32])data)[yte(hash[:], []b
	copy(yte{}= [32]b)
	hash :.UnixNano()Now() time.),ng(StricKey()..Publi signer", tx,%d%v%sSprintf("mt.	data := f.Hash {
Key) typesivatepto.Prcrysigner ace{}, erfh(tx interateTxHas) gentemanceTestSuiuite *Perfor(s}

func 	}
), 1000)
r.PublicKey(useutation(rRepselizeUtance.Initia.daoIns		suite

err)t, rror(	require.NoE	000)
), 10Key(.Publicuserkens(tTonstance.Minite.daoIerr := su
		rs {nge use:= rauser or _, eKey) {
	fpto.Privat.cry, users ..ing.Tt *testrs(stUseetupTeTestSuite) srmanceite *Perfonc (suhods

fuHelper met
// }
	}
}
	workers)
	",  workersfor %d%% t 50d be at leasncy shoulieng effic.0, "Scali 50fficiency,ater(t, eassert.Gre8 {
			 <= rs worke	ifers)
	 8 work for up to 50%least(at nable asould be recy sho/ Efficien		/y)

efficienc workers, y: %.1f%%",g efficienc Scalin%d,"Workers: 	t.Logf(* 100

	4(duration) loat6edTime) / fect(exp:= float64ency ficis)
		efkerration(wor time.Du /Timebase= me :	expectedTi

			}ontinue
		c {
	ers == 1works {
		if ltresuange  := rtionra durs,r workes[1]
	foe := resultim	baseTficiency
ling efyze scaal// An

	
	}hput)ougation, throrkers, dur wec",f ops/shput: %.2ougn: %v, Thr %d, Duratioers:gf("Work

		t.Lods()tion.Secon) / duraotalOps4(tat6= flot :ghpu	throu
	ernsPerWorkatiors * oper= workeotalOps :		tion

 = duratorkers][wsults
		retart)e(sime.Sinc:= t	duration ait()
	.W	wg		}

	}(w)

			
				}roposalHash)cKey(), pbliTx, user.Pusaltion(propoacDAOTranscessnce.ProdaoInstaite.			suuser)
		osalTx, (propateTxHashersuite.gensh := posalHa	pro

					}
				ndomHash(),: suite.raHashta			Metada	0,
		ld:    100		Thresho		00,
		 + 36Unix()me.Now().   tidTime:   		En
				() - 100,Now().Unix time.Time:   		Start				
e,SimplpengTy dao.Votie:  Typ			Voting		l,
	peGenera.ProposalType: dao	ProposalTy					s",
ckene bottlurrencyTesting conc "on: 	Descripti					D, i),
, workerIest %d-%d" T"Concurrency.Sprintf(    fmt    	Title:				    200,
	Fee:      			lTx{
			oposa := &dao.PrsalTx				propoons
	xed operatiMi					// ]

*10+i%10s[workerIDser := user				u	++ {
Worker; iperationsPer i < o:= 0;			for i 

	ne() wg.Do				defer
ID int) {unc(worker		go f(1)
	wg.Add			; w++ {
 w < workers	for w := 0;
	aitGroup
g sync.W
		var wow()rt := time.N

		sta)ers...tUsers(t, ussetupTes	suite.
		}
	teKey()neratePriva crypto.Gers[i] =se {
			ue users := rang		for irs*10)
workeateKey, pto.Priv[]cry := make(		usersis test
 for thtup users{
		// SemWorkers = range nukers : wor	for _,ration)

[int]time.Du= make(map :lts

	resu= 100 :onsPerWorkererati, 16}
	op, 8{1, 2, 4s := []int	numWorkerks
ttleneccy borent for concur
	// Tesg.T) {ks(t *testinnecrrencyBottletestConcute) anceTestSui *Performtefunc (sui

)
}10ms"han ess td be lse shoulauge GC p "Averallisecond,*time.MiPause, 10gGC.Less(t, av	assertcles)
Cyion(gcDurat/ time.e) ausDuration(gcP:= time.se PauvgGCs
	aertionC impact ass	// G))

(gcCyclesDurationtime.on(gcPause)/.Durati", timee: %vC paus("Average G	t.Logf)
ause)Duration(gcP %v", time.se:GC pauotal gf("T	t.Los)
Cycle %d", gccles:C cygf("G	t.Loon)
ns, duratimOperatio", nuvrations in %: %d ope testimpact"GC Logf(

	t.Nstal m1.PauseTolNs -m2.PauseTota= 	gcPause : m1.NumGC
= m2.NumGC -cCycles :	g
m2)
ats(&eadMemSttime.Rart)
	runSince(st:= time.n atio

	dur		}
	})
C(.Gruntime
			 0 {i%100 ==		if ions
aterery 100 opForce GC ev		// 

ash), proposalHicKey() user.PubloposalTx,ansaction(prsDAOTrtance.Proceste.daoIns	sui)
	userposalTx, (proxHashe.generateTuitsh := s	proposalHa

	(),
		}domHashte.ran suish:ataHa			Metad000,
  1eshold:  	Thr00,
		Unix() + 36w().ime.No   te:   EndTim			x() - 100,
e.Now().Uniime:    timartT	St		ple,
TypeSimo.VotingngType:   daoti	Vl,
		eraTypeGenao.Proposalype: d	ProposalT		act",
 impsting GC"Tetion:  crip
			Desd", i), %osalropC Test PSprintf("G     fmt.tle:   		Ti	     200,
    		Fee: osalTx{
	opdao.Pr &Tx :=		proposal GC
iggerhat will trcts tbjemporary o Create te	//

	%len(users)]:= users[i	user ; i++ {
	nsatio< numOperr i := 0; i 
	fo00
 10:=ions perat()
	numO := time.Now	start GC
e monitoring whilnsrm operatiofo// Per...)

	ersUsers(t, usetupTest.ste
	}
	suieKey()ivateratePrrypto.Geners[i] = c{
		use users or i := rang, 200)
	f.PrivateKey]cryptors := make([up
	use

	// Settats(&m1)e.ReadMemS
	runtimtatsntime.MemS2 rum1, me
	var mancforpact on perst GC im	// Te
.T) {t *testingonImpact(bageCollectie) testGareTestSuitPerformancte *unc (sui")
}

femory of setup mss than halfould be leory shion memerat, "Op/2rysetupMemoy, onMemoroperatit.Less(t, 
	asserrtionsiency assery effic	// Memo

y))emor4(setupMt6Memory)/floa64(operationfloat%.2f", ency ratio: emory effici"MLogf(ry)
	t.erationMemoes", opmory: %d bytn me"Operatio
	t.Logf(tupMemory)", seytesmory: %d b"Setup me(t.Logfoc

	- m2.Allc lo m3.AlMemory :=onperati1.Alloc
	oc - mAllo2.Memory := m

	setup(&m3)atsmStme.ReadMe
	runtiGC()ime.ons
	runtperati after o
	// Memory
	}	}
cKey())
	.Publier(userowtiveVotingPEffecstance.GetoInuite.da	s
		PublicKey())user.eputation(etUserRtance.Ge.daoInsit		su	))
icKey(ublce(user.PTokenBalanstance.GetaoIn		suite.dient
	mory-efficuld be meho that s/ Operations	
			/	sers)]
	(ulenrs[i%ser := use
			u; i++ {< 1000; i or i := 		f
d++ {d < 10; rounun := 0; roor roundmory
	f meaccumulatet ould not shns thany operatioPerform ma)

	// ts(&m2.ReadMemSta
	runtimee.GC()imunt	rter setup
ory afMem

	// t, users...)ers(tUse.setupTes	}
	suiteKey()
atePrivypto.Generatrs[i] = cr
		usenge users { := ra
	for iey, 100)to.PrivateKmake([]cryp= 
	users :onsperati Perform o

	//ats(&m1)dMemSt	runtime.Reae.GC()

	runtime memory Baselin

	//.MemStatsruntime, m3  m2ar m1,ns
	voperatioth repeated ficiency wist memory ef Te	//sting.T) {
tet *ciency(oryEffiite) testMemmanceTestSue *Perfor(suit}

func zation")
 CPU optimith/sec wiperations 1000 oeast at lieveShould ach.0, ", 1000hroughputeater(t, t	assert.Grtions
ssernce a	// Performa)

putghouthrn, ns, duratiotalOperatio)", to2f ops/secs in %v (%.operationest: %d ization tU util("CPogf

	t.Lon.Seconds()s) / duratiationotalOperfloat64(tput := 
	throughkerrWorrationsPeers * numOpeumWorkations := nper	totalOrt)

ce(staSinime.n := tatiourt()
	d
	wg.Wai

	}		}(w)

			}		}y())
		PublicKeer(user.gPowVotinivee.GetEffecttancoIns	suite.da				ulation
 power calcingot // V		case 3:		ey())

PublicKtation(user..GetUserRepudaoInstanceuite.		s	
		lculation caReputation 2: // 			case 1)

	ey(),licKey(), to.Pubuser.PublicKsferTokens(ance.TrandaoInstsuite.			ers)]
		))%len(us0+(i+1kerID*1or:= users[(w				to 	ansfer
 Token trase 1: //			ch)

	posalHasprolicKey(), user.PublTx, tion(proposansacssDAOTratance.Procenssuite.daoI					user)
oposalTx, prHash(rateTxenee.g= suitsh :	proposalHa
						}
			mHash(),
.randoaHash: suiteat		Metad				   1000,
reshold: 		Th			
	0,x() + 360.Uniime.Now()   tme:   ndTi			E
			00,) - 1ow().Unix(me.N:    ti	StartTime					Simple,
pedao.VotingTy  ingType: 	Vot					,
eneralalTypeG: dao.ProposypeoposalT				Pr
		n test",U utilizatio:  "CPcription		Des			D, i),
	erIork", wosal %d-%d Test Proprintf("CPUfmt.Sp     	Title:   				 200,
	  ee:       			F	alTx{
		.Propos:= &daoTx osal		proption
			al crea// Propos 0: ase		c{
		4  i % 	switch			ons
 of operati		// Mix

		0]i%1kerID*10+[worusersser := {
				u++ ; insPerWorkerOperatio= 0; i < numr i :	fo

		g.Done()er w			defint) {
ID worker		go func(dd(1)
wg.A
		++ {rkers; ww < numWo= 0; for w :ions
	rent operaturive conc CPU-intens//up

	aitGroar wg sync.W.Now()
	vmetart := ti)

	s, users...tUsers(tTes	suite.setup
	}
ateKey()eneratePriv= crypto.G] sers[i		u{
ange users := rfor i 10)
	umWorkers*teKey, nto.Privamake([]cryp	users := etup users
	// S

s)umWorkerrkers", nwowith %d n atioU utilizTesting CPf(".Log000

	tWorker := 1rationsPernumOpePU()
	mCe.NuruntimmWorkers := test
	nu operations tensivein{
	// CPU-T) (t *testing.UtilizationestCPUuite) tormanceTestS *Perf
func (suitests
 te utilization// Resourceal")
}

50MB totss than hould use le), "S0*1024*1024int64(5d, uemoryUse, mess(t
	assert.Luser")er  10KB pless thanhould use  "St64(10000),PerUser, uinemory.Less(t, mns
	asserttioassery usage / Memor
	/umGC)
GC-m1.N m2.Numycles: %d",.Logf("GC c	tjects)
m2.HeapObts: %d", p objec("Heaogf
	t.LerUser)yP memor memoryUsed,r",r useytes pe %d btotal,tes  bysage: %df("Memory u
	t.LogumUsers)
 / uint64(ned := memoryUserUser
	memoryPllocAlloc - m1.Aed := m2.emoryUs

	mStats(&m2)ReadMem	runtime.()
me.GC	runti	}

salHash)
ey(), propoor.PublicK, creatposalTxtion(proAOTransacssDProceance.te.daoInstor)
		suiatposalTx, cre(proteTxHashsuite.genera=  :posalHash
		pro

		}(),omHashandte.rsuidataHash: 	Meta		:    1000,
	Threshold600,
		 3.Unix() +time.Now()    	EndTime:  		 100,
() -).Unixtime.Now(ime:    			StartT,
ypeSimpleao.VotingT:   dpegTytin		Voneral,
	TypeGeosaldao.PropalType: 			Propos, i),
d"l %th proposa wi usageemoryting m("Tesrintfion:  fmt.SpDescript		
	, %d", i)alos Propmory TestMe"ntf(  fmt.Spri      Title:		   200,
	:       			FeeoposalTx{
x := &dao.Pr		proposalTumUsers]
rs[i%nusecreator :=  {
		0; i++ i < 50r i := 0;	foroposals
 pCreate many

	// , users...)ers(tTestUs.setup	suite
	}
ateKey()GeneratePrivpto.s[i] = cry	user	rs {
= range use
	for i :ers)numUsKey, ypto.Privateke([]crmas := 0
	user= 100s :
	numUsernsratioe opesivintenry-orm memoerf
	// Pats(&m1)
mStMeuntime.ReadGC()
	rme.untis
	rime.MemStatnt m1, m2 runs
	varoperatioge during memory usaonitor  {
	// M *testing.T)nderLoad(tmoryUsageUtMete) tesestSuieT *Performancc (suite
fun")
}
ariouency scen high-freqsec intes/ast 500 vot lehieve auld acShot, 500.0, "throughpueater(t, t.Gr
	assernsce assertiorman/ Perfo

	/otes))
	}i, len(v votes",  received %dal %dgf("Propos	t.Lo
	, err)Error(tuire.No	req	salHash)
ropootes(pce.GetVstanIn.dao := suitees, errothes {
		vproposalHasange  rosalHash :=r i, prop	foded
e recorfy votes wereri	// V

roughput)ration, th, duotesalVsec)", tots/2f votees in %v (%.eted: %d voting complcy vot-frequenLogf("Highs()

	t.Secondion./ duratlVotes) 4(tota= float6throughput :oposals
	mPr * numVoterstes := nutalVort)

	tostae.Since(imon := turatiait()
	d}

	wg.W}(voter)
			}
		teHash)
	icKey(), vo v.Publtion(voteTx,ssDAOTransacrocestance.P.daoInite	su
			x, v)Hash(voteTnerateTxite.ge:= su		voteHash 

		,
				}e" voth frequency   "Higson:  				Rea  10,
		Weight:   		e,
		   choic:  		Choice			sh,
osalHaropalID: popos				Pr 100,
	      	Fee: 				eTx{
.Vot= &daooteTx :
				v			}

	NooteChoice= dao.Ve 	choic				 == 0 {
		if i%2eYes
		.VoteChoichoice := dao				ces {
salHashrange propoHash := , proposal		for i)

	one(.Dr wg
			defe {eKey).Privatc(v cryptofun
		go d(1)Adwg.ers {
		nge vot voter := ra
	for _,tGroupsync.Wai

	var wg ime.Now():= t	start ls)
osaPropoters, numnumV", lsosaroprs on %d p: %d voteency votingqu high-fref("StartingLog
	t.y votinggh-frequenc}

	// Hi
	oposalHash] = pralHashes[i		propos

)err(t, uire.NoError
		reqosalHash)), propey(PublicK creator.roposalTx,ion(pDAOTransactProcessance..daoInst= suite		err :eator)
osalTx, crxHash(propenerateTte.g:= suiroposalHash 	}

		p
	),ash(e.randomHaHash: suit		Metadat,
	000d:    1Threshol			 3600,
).Unix() +   time.Now(   EndTime:		100,
	).Unix() - ime.Now(tTime:    t			StarSimple,
pegTy dao.Votin  ngType:Votil,
			raTypeGeneao.Proposale: dTypposalPro			,
ng test"uency votiHigh freqiption:  "
			Descr i),al %d",Proposncy Frequeh ("Higmt.Sprintf       f: itle	T200,
		:          			FeeoposalTx{
o.Pr:= &dasalTx {
		propoi++ osals;  < numProp; ii := 0
	for osals)
ash, numProppes.H:= make([]tyes osalHash 10
	proposals :=
	numProp proposalstipleul/ Create m
	/)...)
orers, creat(vots(t, appenderetupTestUsite.s	}
	su()
KeyateatePrivnerypto.Ges[i] = cr{
		voterters range voor i := 	f)
tersey, numVoivateKPre([]crypto.= mak	voters : 1000
Voters :=y()
	numePrivateKe.Generatcryptocreator := 
	voting testrequency  for high-f
	// Setup.T) {sting *teng(ttiFrequencyVotHighite) tesanceTestSu*Performte uinc (sfus")
}

in 10 secondte withleompshould calculations  "Power cme.Second,ation, 10*tiulationDur(t, calcassert.Lessnds")
	 30 secote withinmpleup should cogation set"Deleond, .Secimen, 30*ttupDuratioss(t, se
	assert.Leonsassertie ncma
	// Perforer)
EffectivePower, totaledPowelegat maxDwer: %d",ctive poeffel  Totapower: %d, delegated "Maxogf(tion)
	t.LculationDuracal%v",  in ation powersleg de"Calculated.Logf(art)
	tce(st time.Sinn :=ationDuratioalcul
	c
	}
		}tedPower
r = delegaedPowe	maxDelegat
		ower {xDelegatedPPower > ma delegatedif
		ivePower effectwer +=EffectivePototal			
	)
cKey()r.Publiwer(useegatedPoDelInstance.Gete.dao := suitatedPower
		deleglicKey())Pub(user.VotingPowertEffective.GedaoInstancee. := suittivePower	effec	s {
erge us := ranor _, user
	f0)
int64(= utivePower :	totalEffec0)
nt64(er := uiPowxDelegated

	mame.Now()tistart = tions
	wer calculation po Test delega
	//ion)
 setupDuratv",n %ns ition chait up delega	t.Logf("Se
start) time.Since( :=ionaturpD
	setu
	}
ror(t, err)quire.NoEr		reash)
 delegationH),y(KelicPubi].users[, ationTxaction(delegAOTransocessDance.Pr.daoInsterr := suites[i])
		 usertionTx,sh(delegaHa.generateTxsh := suitedelegationHa

		}alse,
		:   fevoke00,
			Rtion: 864ra			Du delegate,
		Delegate:
	   200,:   	Fee
		legationTx{ := &dao.DeegationTx

		delnue
		}			contiion
legatSkip de/  /		case 3:)
y(icKers].Publ%numUse7)i* users[( =	delegate		ion
legatRandom dese 2: // 	ca()
	ublicKey].P users[0 =delegate			
r 0) to uselegatetern (deat 1: // Hub p		caseKey()
blic.Pus[i+1] = userdelegatein
			ar cha// Line	case 0:  % 4 {
	 iswitchterns
		ation patent deleg differ/ Create		
		/ey
ypto.PublicKlegate cr	var de++ {
	Users-1; i; i < num	for i := 0
d networkshains an c delegationte
	// Crea.Now()
= timestart :umUsers)
	 ns",userd s with %tion chainplex delegaomg ceatin	t.Logf("Cr...)

ers(t, usersestUsite.setupT	su
	}
teKey()ePrivaratto.Genes[i] = cryp		usersers {
 := range u ior	frs)
 numUse.PrivateKey,cryptos := make([]0
	user100numUsers := tworks
	elegation neomplex date c{
	// Cresting.T) ins(t *teationChamplexDelegCoite) teststSuceTeane *Performit
func (suale")
}
ec at scposals/seast 20 protain at lhould main, 20.0, "S throughputr(t,ert.Greateassns
	assertiomance 	// Perforosals)

numProposals), ropt, len(paterOrEqual(sert.Gre	asls()
oposalPr.ListAl.daoInstancels := suiteosad
	prop were createy proposals	// Verif
ut)
ughproion, th, duratosalsPropec)", numposals/sprov (%.2f  in %salsropoed %d pogf("Creat.L	t

()on.Secondsurati) / dposals(numPro= float64ughput :hroart)
	tince(stn := time.Sduratio	}

			}
rate)
i+1, )", als/secosrop%.2f pposals ( %d proCreated("		t.Logf)
	.Seconds(elapsed4(i+1) / oat6= fl		rate :	start)
Since(time.= sed :ap{
			el1000 == 0 )% (i+1		ifsals
poy 1000 proogress ever pr
		// Log
r)or(t, erquire.NoErrh)
		resalHasey(), propotor.PublicKlTx, creaoposaprction(nsaDAOTrance.Processte.daoInsta	err := suieator)
	osalTx, cr(propHashteTxnera suite.geosalHash :=prop}

		
		omHash(),.randteaHash: sui		Metadat
	ld:    1000,shore
			Th600,nix() + 3Now().Uime. t     	EndTime:		0,
nix() - 10ime.Now().U time:   artT
			StgTypeSimple,o.Votinype:   da			VotingTeral,
TypeGenosaldao.PropsalType: 
			Propo%d", i),osal ty test propalabiliSc"tf(Sprinn:  fmt.iptio		Descr i),
	posal %d",f("Mass Prorint fmt.Sp     e:  
			Titl      200,   e: 
			FeosalTx{dao.Propx := &proposalT]

		numCreatorsreators[i%= cator :	crei++ {
	sals; i < numPropor i := 0; ()

	fotime.Nowtart := osals)
	sPropals", numroposeating %d p"Cr
	t.Logf(...)
rsators(t, creupTestUsee.set	suity()
	}
ePrivateKepto.Generatcryators[i] = 
		cres {ange creator i := rs)
	for, numCreatorvateKeyPri]crypto.= make([ors :	creat

= 50eators :mCr000
	nu 5als :=ropos	numP proposals
umber ofassive nt with m	// Tesg.T) {
tinolume(t *tesposalVassiveProte) testMuieTestSformanc*Perite nc (su

fus")
}condin 30 see withomplet should c"Operationsond, Sectime.n, 30*DuratiotionperaLess(t, o	assert.nds")
hin 60 secoe witomplet cldetup shou"User sd, 0*time.Seconration, 6, setupDuss(tsert.Lertions
	assence asrforma)

	// PeonDurations, operati", numUserrs in %v on %d useationserrmed op"Perfot.Logf((start)
	Since := time.ionDurationrat
	ope()
vityDecayyInactince.Appl.daoInsta
	suitedecay reputation 
	// Applysers)
), numUing, len(rankl(tuaEqreaterOr
	assert.Ging()ationRankGetReputnstance.ite.daoIking := suing
	rannkon rautatiGet rep

	// w()time.Nort = sta base
	 large userwithrations st ope

	// Teion)Durat setupers,numUs, s in %v"user"Set up %d 
	t.Logf(start)e.Since(tion := tim
	setupDura	}
	}
))
	uint64(100+iey(), i].PublicKrs[putation(useUserRealizeance.Initinstsuite.daoI		putation
	lize re	// Initia		
rr)
Error(t, eequire.Noi))
			r64(1000+ntlicKey(), uiPub(users[i].nsntTokestance.MioIn.date sui		err :=okens
		// Mint t	
		Key()
		teeneratePriva = crypto.G		users[i]
	ge users {= ran	for i :	tchSize)
Key, bao.Privateke([]cryptsers := ma		uch++ {
; batizetchSnumUsers/ba< 0; batch  := r batch= 100
	foSize :tchba	ues
ssmory id mehes to avoiatce users in beat
	// Crme.Now()
tart := ti	smUsers)

s", nu%d usering with Logf("Testt.0
	500s := numUser users
	e number ofargwith l
	// Test ting.T) {t *tesgeUserBase(testLarSuite) manceTest *Perfor(suiteunc tests

fity calabil// S

conds")
} se10 within  completens shouldioixed operatSecond, "M0*time. duration, 1t,.Less(ssertation)
	adur: %v", eted intions complerarent opixed concurt.Logf("M

	ince(start)on := time.Srati	du
	wg.Wait()()

		}
	}ash)
elegationHblicKey(), dgator.PuonTx, deleion(delegatiransactcessDAOTstance.Pro.daoInsuite
			ator)onTx, delegatish(delegeTxHaeratgenite. sush :=ationHa			deleg			}

:   false,
			Revoke 86400,
		Duration:(),
			PublicKeyelegate.legate: d			De,
	:      200	Fee			gationTx{
o.Dele
