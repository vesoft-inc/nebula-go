package nebula

type ValueBuilder struct {
	obj *Value
}

func NewValueBuilder() *ValueBuilder {
	return &ValueBuilder{
		obj: NewValue(),
	}
}

func (p ValueBuilder) Emit() *Value {
	return &Value{
		NVal:  p.obj.NVal,
		BVal:  p.obj.BVal,
		IVal:  p.obj.IVal,
		FVal:  p.obj.FVal,
		SVal:  p.obj.SVal,
		DVal:  p.obj.DVal,
		TVal:  p.obj.TVal,
		DtVal: p.obj.DtVal,
		VVal:  p.obj.VVal,
		EVal:  p.obj.EVal,
		PVal:  p.obj.PVal,
		LVal:  p.obj.LVal,
		MVal:  p.obj.MVal,
		UVal:  p.obj.UVal,
		GVal:  p.obj.GVal,
		GgVal: p.obj.GgVal,
		DuVal: p.obj.DuVal,
	}
}

func (v *ValueBuilder) NVal(nVal *NullType) *ValueBuilder {
	v.obj.NVal = nVal
	return v
}

func (v *ValueBuilder) BVal(bVal *bool) *ValueBuilder {
	v.obj.BVal = bVal
	return v
}

func (v *ValueBuilder) IVal(iVal *int64) *ValueBuilder {
	v.obj.IVal = iVal
	return v
}

func (v *ValueBuilder) FVal(fVal *float64) *ValueBuilder {
	v.obj.FVal = fVal
	return v
}

func (v *ValueBuilder) SVal(sVal []byte) *ValueBuilder {
	v.obj.SVal = sVal
	return v
}

func (v *ValueBuilder) DVal(dVal *Date) *ValueBuilder {
	v.obj.DVal = dVal
	return v
}

func (v *ValueBuilder) TVal(tVal *Time) *ValueBuilder {
	v.obj.TVal = tVal
	return v
}

func (v *ValueBuilder) DtVal(dtVal *DateTime) *ValueBuilder {
	v.obj.DtVal = dtVal
	return v
}

func (v *ValueBuilder) VVal(vVal *Vertex) *ValueBuilder {
	v.obj.VVal = vVal
	return v
}

func (v *ValueBuilder) EVal(eVal *Edge) *ValueBuilder {
	v.obj.EVal = eVal
	return v
}

func (v *ValueBuilder) PVal(pVal *Path) *ValueBuilder {
	v.obj.PVal = pVal
	return v
}

func (v *ValueBuilder) LVal(lVal *NList) *ValueBuilder {
	v.obj.LVal = lVal
	return v
}

func (v *ValueBuilder) MVal(mVal *NMap) *ValueBuilder {
	v.obj.MVal = mVal
	return v
}

func (v *ValueBuilder) UVal(uVal *NSet) *ValueBuilder {
	v.obj.UVal = uVal
	return v
}

func (v *ValueBuilder) GVal(gVal *DataSet) *ValueBuilder {
	v.obj.GVal = gVal
	return v
}

func (v *ValueBuilder) GgVal(ggVal *Geography) *ValueBuilder {
	v.obj.GgVal = ggVal
	return v
}

func (v *ValueBuilder) DuVal(duVal *Duration) *ValueBuilder {
	v.obj.DuVal = duVal
	return v
}

func (v *Value) SetNVal(nVal *NullType) *Value {
	v.NVal = nVal
	return v
}

func (v *Value) SetBVal(bVal *bool) *Value {
	v.BVal = bVal
	return v
}

func (v *Value) SetIVal(iVal *int64) *Value {
	v.IVal = iVal
	return v
}

func (v *Value) SetFVal(fVal *float64) *Value {
	v.FVal = fVal
	return v
}

func (v *Value) SetSVal(sVal []byte) *Value {
	v.SVal = sVal
	return v
}

func (v *Value) SetDVal(dVal *Date) *Value {
	v.DVal = dVal
	return v
}

func (v *Value) SetTVal(tVal *Time) *Value {
	v.TVal = tVal
	return v
}

func (v *Value) SetDtVal(dtVal *DateTime) *Value {
	v.DtVal = dtVal
	return v
}

func (v *Value) SetVVal(vVal *Vertex) *Value {
	v.VVal = vVal
	return v
}

func (v *Value) SetEVal(eVal *Edge) *Value {
	v.EVal = eVal
	return v
}

func (v *Value) SetPVal(pVal *Path) *Value {
	v.PVal = pVal
	return v
}

func (v *Value) SetLVal(lVal *NList) *Value {
	v.LVal = lVal
	return v
}

func (v *Value) SetMVal(mVal *NMap) *Value {
	v.MVal = mVal
	return v
}

func (v *Value) SetUVal(uVal *NSet) *Value {
	v.UVal = uVal
	return v
}

func (v *Value) SetGVal(gVal *DataSet) *Value {
	v.GVal = gVal
	return v
}

func (v *Value) SetGgVal(ggVal *Geography) *Value {
	v.GgVal = ggVal
	return v
}

func (v *Value) SetDuVal(duVal *Duration) *Value {
	v.DuVal = duVal
	return v
}

func (n *NList) SetValues(values []*Value) *NList {
	n.Values = values
	return n
}

type NListBuilder struct {
	obj *NList
}

func NewNListBuilder() *NListBuilder {
	return &NListBuilder{
		obj: NewNList(),
	}
}

func (p NListBuilder) Emit() *NList {
	return &NList{
		Values: p.obj.Values,
	}
}

func (n *NListBuilder) Values(values []*Value) *NListBuilder {
	n.obj.Values = values
	return n
}

type NMapBuilder struct {
	obj *NMap
}

func NewNMapBuilder() *NMapBuilder {
	return &NMapBuilder{
		obj: NewNMap(),
	}
}

func (p NMapBuilder) Emit() *NMap {
	return &NMap{
		Kvs: p.obj.Kvs,
	}
}

func (n *NMapBuilder) Kvs(kvs map[string]*Value) *NMapBuilder {
	n.obj.Kvs = kvs
	return n
}

func (n *NMap) SetKvs(kvs map[string]*Value) *NMap {
	n.Kvs = kvs
	return n
}

type NSetBuilder struct {
	obj *NSet
}

func NewNSetBuilder() *NSetBuilder {
	return &NSetBuilder{
		obj: NewNSet(),
	}
}

func (p NSetBuilder) Emit() *NSet {
	return &NSet{
		Values: p.obj.Values,
	}
}

func (n *NSetBuilder) Values(values []*Value) *NSetBuilder {
	n.obj.Values = values
	return n
}

func (n *NSet) SetValues(values []*Value) *NSet {
	n.Values = values
	return n
}

type RowBuilder struct {
	obj *Row
}

func NewRowBuilder() *RowBuilder {
	return &RowBuilder{
		obj: NewRow(),
	}
}

func (p RowBuilder) Emit() *Row {
	return &Row{
		Values: p.obj.Values,
	}
}

func (r *RowBuilder) Values(values []*Value) *RowBuilder {
	r.obj.Values = values
	return r
}

func (r *Row) SetValues(values []*Value) *Row {
	r.Values = values
	return r
}

type DataSetBuilder struct {
	obj *DataSet
}

func NewDataSetBuilder() *DataSetBuilder {
	return &DataSetBuilder{
		obj: NewDataSet(),
	}
}

func (p DataSetBuilder) Emit() *DataSet {
	return &DataSet{
		ColumnNames: p.obj.ColumnNames,
		Rows:        p.obj.Rows,
	}
}

func (d *DataSetBuilder) ColumnNames(columnNames [][]byte) *DataSetBuilder {
	d.obj.ColumnNames = columnNames
	return d
}

func (d *DataSetBuilder) Rows(rows []*Row) *DataSetBuilder {
	d.obj.Rows = rows
	return d
}

func (d *DataSet) SetColumnNames(columnNames [][]byte) *DataSet {
	d.ColumnNames = columnNames
	return d
}

func (d *DataSet) SetRows(rows []*Row) *DataSet {
	d.Rows = rows
	return d
}

type CoordinateBuilder struct {
	obj *Coordinate
}

func NewCoordinateBuilder() *CoordinateBuilder {
	return &CoordinateBuilder{
		obj: NewCoordinate(),
	}
}

func (p CoordinateBuilder) Emit() *Coordinate {
	return &Coordinate{
		X: p.obj.X,
		Y: p.obj.Y,
	}
}

func (c *CoordinateBuilder) X(x float64) *CoordinateBuilder {
	c.obj.X = x
	return c
}

func (c *CoordinateBuilder) Y(y float64) *CoordinateBuilder {
	c.obj.Y = y
	return c
}

func (c *Coordinate) SetX(x float64) *Coordinate {
	c.X = x
	return c
}

func (c *Coordinate) SetY(y float64) *Coordinate {
	c.Y = y
	return c
}

type PointBuilder struct {
	obj *Point
}

func NewPointBuilder() *PointBuilder {
	return &PointBuilder{
		obj: NewPoint(),
	}
}

func (p PointBuilder) Emit() *Point {
	return &Point{
		Coord: p.obj.Coord,
	}
}

func (p *PointBuilder) Coord(coord *Coordinate) *PointBuilder {
	p.obj.Coord = coord
	return p
}

func (p *Point) SetCoord(coord *Coordinate) *Point {
	p.Coord = coord
	return p
}

type PolygonBuilder struct {
	obj *Polygon
}

func NewPolygonBuilder() *PolygonBuilder {
	return &PolygonBuilder{
		obj: NewPolygon(),
	}
}

func (p PolygonBuilder) Emit() *Polygon {
	return &Polygon{
		CoordListList: p.obj.CoordListList,
	}
}

func (p *PolygonBuilder) CoordListList(coordListList [][]*Coordinate) *PolygonBuilder {
	p.obj.CoordListList = coordListList
	return p
}

func (p *Polygon) SetCoordListList(coordListList [][]*Coordinate) *Polygon {
	p.CoordListList = coordListList
	return p
}

type GeographyBuilder struct {
	obj *Geography
}

func NewGeographyBuilder() *GeographyBuilder {
	return &GeographyBuilder{
		obj: NewGeography(),
	}
}

func (p GeographyBuilder) Emit() *Geography {
	return &Geography{
		PtVal: p.obj.PtVal,
		LsVal: p.obj.LsVal,
		PgVal: p.obj.PgVal,
	}
}

func (g *GeographyBuilder) PtVal(ptVal *Point) *GeographyBuilder {
	g.obj.PtVal = ptVal
	return g
}

func (g *GeographyBuilder) LsVal(lsVal *LineString) *GeographyBuilder {
	g.obj.LsVal = lsVal
	return g
}

func (g *GeographyBuilder) PgVal(pgVal *Polygon) *GeographyBuilder {
	g.obj.PgVal = pgVal
	return g
}

func (g *Geography) SetPtVal(ptVal *Point) *Geography {
	g.PtVal = ptVal
	return g
}

func (g *Geography) SetLsVal(lsVal *LineString) *Geography {
	g.LsVal = lsVal
	return g
}

func (g *Geography) SetPgVal(pgVal *Polygon) *Geography {
	g.PgVal = pgVal
	return g
}

type TagBuilder struct {
	obj *Tag
}

func NewTagBuilder() *TagBuilder {
	return &TagBuilder{
		obj: NewTag(),
	}
}

func (p TagBuilder) Emit() *Tag {
	return &Tag{
		Name:  p.obj.Name,
		Props: p.obj.Props,
	}
}

func (t *TagBuilder) Name(name []byte) *TagBuilder {
	t.obj.Name = name
	return t
}

func (t *TagBuilder) Props(props map[string]*Value) *TagBuilder {
	t.obj.Props = props
	return t
}

func (t *Tag) SetName(name []byte) *Tag {
	t.Name = name
	return t
}

func (t *Tag) SetProps(props map[string]*Value) *Tag {
	t.Props = props
	return t
}

type VertexBuilder struct {
	obj *Vertex
}

func NewVertexBuilder() *VertexBuilder {
	return &VertexBuilder{
		obj: NewVertex(),
	}
}

func (p VertexBuilder) Emit() *Vertex {
	return &Vertex{
		Vid:  p.obj.Vid,
		Tags: p.obj.Tags,
	}
}

func (v *VertexBuilder) Vid(vid *Value) *VertexBuilder {
	v.obj.Vid = vid
	return v
}

func (v *VertexBuilder) Tags(tags []*Tag) *VertexBuilder {
	v.obj.Tags = tags
	return v
}

func (v *Vertex) SetVid(vid *Value) *Vertex {
	v.Vid = vid
	return v
}

func (v *Vertex) SetTags(tags []*Tag) *Vertex {
	v.Tags = tags
	return v
}

type EdgeBuilder struct {
	obj *Edge
}

func NewEdgeBuilder() *EdgeBuilder {
	return &EdgeBuilder{
		obj: NewEdge(),
	}
}

func (p EdgeBuilder) Emit() *Edge {
	return &Edge{
		Src:     p.obj.Src,
		Dst:     p.obj.Dst,
		Type:    p.obj.Type,
		Name:    p.obj.Name,
		Ranking: p.obj.Ranking,
		Props:   p.obj.Props,
	}
}

func (e *EdgeBuilder) Src(src *Value) *EdgeBuilder {
	e.obj.Src = src
	return e
}

func (e *EdgeBuilder) Dst(dst *Value) *EdgeBuilder {
	e.obj.Dst = dst
	return e
}

func (e *EdgeBuilder) Type(type_a1 EdgeType) *EdgeBuilder {
	e.obj.Type = type_a1
	return e
}

func (e *EdgeBuilder) Name(name []byte) *EdgeBuilder {
	e.obj.Name = name
	return e
}

func (e *EdgeBuilder) Ranking(ranking EdgeRanking) *EdgeBuilder {
	e.obj.Ranking = ranking
	return e
}

func (e *EdgeBuilder) Props(props map[string]*Value) *EdgeBuilder {
	e.obj.Props = props
	return e
}

func (e *Edge) SetSrc(src *Value) *Edge {
	e.Src = src
	return e
}

func (e *Edge) SetDst(dst *Value) *Edge {
	e.Dst = dst
	return e
}

func (e *Edge) SetType(type_a1 EdgeType) *Edge {
	e.Type = type_a1
	return e
}

func (e *Edge) SetName(name []byte) *Edge {
	e.Name = name
	return e
}

func (e *Edge) SetRanking(ranking EdgeRanking) *Edge {
	e.Ranking = ranking
	return e
}

func (e *Edge) SetProps(props map[string]*Value) *Edge {
	e.Props = props
	return e
}

type StepBuilder struct {
	obj *Step
}

func NewStepBuilder() *StepBuilder {
	return &StepBuilder{
		obj: NewStep(),
	}
}

func (p StepBuilder) Emit() *Step {
	return &Step{
		Dst:     p.obj.Dst,
		Type:    p.obj.Type,
		Name:    p.obj.Name,
		Ranking: p.obj.Ranking,
		Props:   p.obj.Props,
	}
}

func (s *StepBuilder) Dst(dst *Vertex) *StepBuilder {
	s.obj.Dst = dst
	return s
}

func (s *StepBuilder) Type(type_a1 EdgeType) *StepBuilder {
	s.obj.Type = type_a1
	return s
}

func (s *StepBuilder) Name(name []byte) *StepBuilder {
	s.obj.Name = name
	return s
}

func (s *StepBuilder) Ranking(ranking EdgeRanking) *StepBuilder {
	s.obj.Ranking = ranking
	return s
}

func (s *StepBuilder) Props(props map[string]*Value) *StepBuilder {
	s.obj.Props = props
	return s
}

func (s *Step) SetDst(dst *Vertex) *Step {
	s.Dst = dst
	return s
}

func (s *Step) SetType(type_a1 EdgeType) *Step {
	s.Type = type_a1
	return s
}

func (s *Step) SetName(name []byte) *Step {
	s.Name = name
	return s
}

func (s *Step) SetRanking(ranking EdgeRanking) *Step {
	s.Ranking = ranking
	return s
}

func (s *Step) SetProps(props map[string]*Value) *Step {
	s.Props = props
	return s
}

type PathBuilder struct {
	obj *Path
}

func NewPathBuilder() *PathBuilder {
	return &PathBuilder{
		obj: NewPath(),
	}
}

func (p PathBuilder) Emit() *Path {
	return &Path{
		Src:   p.obj.Src,
		Steps: p.obj.Steps,
	}
}

func (p *PathBuilder) Src(src *Vertex) *PathBuilder {
	p.obj.Src = src
	return p
}

func (p *PathBuilder) Steps(steps []*Step) *PathBuilder {
	p.obj.Steps = steps
	return p
}

func (p *Path) SetSrc(src *Vertex) *Path {
	p.Src = src
	return p
}

func (p *Path) SetSteps(steps []*Step) *Path {
	p.Steps = steps
	return p
}

type HostAddrBuilder struct {
	obj *HostAddr
}

func NewHostAddrBuilder() *HostAddrBuilder {
	return &HostAddrBuilder{
		obj: NewHostAddr(),
	}
}

func (p HostAddrBuilder) Emit() *HostAddr {
	return &HostAddr{
		Host: p.obj.Host,
		Port: p.obj.Port,
	}
}

func (h *HostAddrBuilder) Host(host string) *HostAddrBuilder {
	h.obj.Host = host
	return h
}

func (h *HostAddrBuilder) Port(port Port) *HostAddrBuilder {
	h.obj.Port = port
	return h
}

func (h *HostAddr) SetHost(host string) *HostAddr {
	h.Host = host
	return h
}

func (h *HostAddr) SetPort(port Port) *HostAddr {
	h.Port = port
	return h
}

type KeyValueBuilder struct {
	obj *KeyValue
}

func NewKeyValueBuilder() *KeyValueBuilder {
	return &KeyValueBuilder{
		obj: NewKeyValue(),
	}
}

func (p KeyValueBuilder) Emit() *KeyValue {
	return &KeyValue{
		Key:   p.obj.Key,
		Value: p.obj.Value,
	}
}

func (k *KeyValueBuilder) Key(key []byte) *KeyValueBuilder {
	k.obj.Key = key
	return k
}

func (k *KeyValueBuilder) Value(value []byte) *KeyValueBuilder {
	k.obj.Value = value
	return k
}

func (k *KeyValue) SetKey(key []byte) *KeyValue {
	k.Key = key
	return k
}

func (k *KeyValue) SetValue(value []byte) *KeyValue {
	k.Value = value
	return k
}

type DurationBuilder struct {
	obj *Duration
}

func NewDurationBuilder() *DurationBuilder {
	return &DurationBuilder{
		obj: NewDuration(),
	}
}

func (p DurationBuilder) Emit() *Duration {
	return &Duration{
		Seconds:      p.obj.Seconds,
		Microseconds: p.obj.Microseconds,
		Months:       p.obj.Months,
	}
}

func (d *DurationBuilder) Seconds(seconds int64) *DurationBuilder {
	d.obj.Seconds = seconds
	return d
}

func (d *DurationBuilder) Microseconds(microseconds int32) *DurationBuilder {
	d.obj.Microseconds = microseconds
	return d
}

func (d *DurationBuilder) Months(months int32) *DurationBuilder {
	d.obj.Months = months
	return d
}

func (d *Duration) SetSeconds(seconds int64) *Duration {
	d.Seconds = seconds
	return d
}

func (d *Duration) SetMicroseconds(microseconds int32) *Duration {
	d.Microseconds = microseconds
	return d
}

func (d *Duration) SetMonths(months int32) *Duration {
	d.Months = months
	return d
}

type LogInfoBuilder struct {
	obj *LogInfo
}

func NewLogInfoBuilder() *LogInfoBuilder {
	return &LogInfoBuilder{
		obj: NewLogInfo(),
	}
}

func (p LogInfoBuilder) Emit() *LogInfo {
	return &LogInfo{
		LogID:          p.obj.LogID,
		TermID:         p.obj.TermID,
		CommitLogID:    p.obj.CommitLogID,
		CheckpointPath: p.obj.CheckpointPath,
	}
}

func (l *LogInfoBuilder) LogID(logID LogID) *LogInfoBuilder {
	l.obj.LogID = logID
	return l
}

func (l *LogInfoBuilder) TermID(termID TermID) *LogInfoBuilder {
	l.obj.TermID = termID
	return l
}

func (l *LogInfoBuilder) CommitLogID(commitLogID LogID) *LogInfoBuilder {
	l.obj.CommitLogID = commitLogID
	return l
}

func (l *LogInfoBuilder) CheckpointPath(checkpointPath []byte) *LogInfoBuilder {
	l.obj.CheckpointPath = checkpointPath
	return l
}

func (l *LogInfo) SetLogID(logID LogID) *LogInfo {
	l.LogID = logID
	return l
}

func (l *LogInfo) SetTermID(termID TermID) *LogInfo {
	l.TermID = termID
	return l
}

func (l *LogInfo) SetCommitLogID(commitLogID LogID) *LogInfo {
	l.CommitLogID = commitLogID
	return l
}

func (l *LogInfo) SetCheckpointPath(checkpointPath []byte) *LogInfo {
	l.CheckpointPath = checkpointPath
	return l
}

type DirInfoBuilder struct {
	obj *DirInfo
}

func NewDirInfoBuilder() *DirInfoBuilder {
	return &DirInfoBuilder{
		obj: NewDirInfo(),
	}
}

func (p DirInfoBuilder) Emit() *DirInfo {
	return &DirInfo{
		Root: p.obj.Root,
		Data: p.obj.Data,
	}
}

func (d *DirInfoBuilder) Root(root []byte) *DirInfoBuilder {
	d.obj.Root = root
	return d
}

func (d *DirInfoBuilder) Data(data [][]byte) *DirInfoBuilder {
	d.obj.Data = data
	return d
}

func (d *DirInfo) SetRoot(root []byte) *DirInfo {
	d.Root = root
	return d
}

func (d *DirInfo) SetData(data [][]byte) *DirInfo {
	d.Data = data
	return d
}

type CheckpointInfoBuilder struct {
	obj *CheckpointInfo
}

func NewCheckpointInfoBuilder() *CheckpointInfoBuilder {
	return &CheckpointInfoBuilder{
		obj: NewCheckpointInfo(),
	}
}

func (p CheckpointInfoBuilder) Emit() *CheckpointInfo {
	return &CheckpointInfo{
		SpaceID:  p.obj.SpaceID,
		Parts:    p.obj.Parts,
		DataPath: p.obj.DataPath,
	}
}

func (c *CheckpointInfoBuilder) SpaceID(spaceID GraphSpaceID) *CheckpointInfoBuilder {
	c.obj.SpaceID = spaceID
	return c
}

func (c *CheckpointInfoBuilder) Parts(parts map[PartitionID]*LogInfo) *CheckpointInfoBuilder {
	c.obj.Parts = parts
	return c
}

func (c *CheckpointInfoBuilder) DataPath(dataPath []byte) *CheckpointInfoBuilder {
	c.obj.DataPath = dataPath
	return c
}

func (c *CheckpointInfo) SetSpaceID(spaceID GraphSpaceID) *CheckpointInfo {
	c.SpaceID = spaceID
	return c
}

func (c *CheckpointInfo) SetParts(parts map[PartitionID]*LogInfo) *CheckpointInfo {
	c.Parts = parts
	return c
}

func (c *CheckpointInfo) SetDataPath(dataPath []byte) *CheckpointInfo {
	c.DataPath = dataPath
	return c
}

type LogEntryBuilder struct {
	obj *LogEntry
}

func NewLogEntryBuilder() *LogEntryBuilder {
	return &LogEntryBuilder{
		obj: NewLogEntry(),
	}
}

func (p LogEntryBuilder) Emit() *LogEntry {
	return &LogEntry{
		Cluster: p.obj.Cluster,
		LogStr:  p.obj.LogStr,
	}
}

func (l *LogEntryBuilder) Cluster(cluster ClusterID) *LogEntryBuilder {
	l.obj.Cluster = cluster
	return l
}

func (l *LogEntryBuilder) LogStr(logStr []byte) *LogEntryBuilder {
	l.obj.LogStr = logStr
	return l
}

func (l *LogEntry) SetCluster(cluster ClusterID) *LogEntry {
	l.Cluster = cluster
	return l
}

func (l *LogEntry) SetLogStr(logStr []byte) *LogEntry {
	l.LogStr = logStr
	return l
}
