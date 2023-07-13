<form name="ConfigureSettings" id="EditView" method="POST">
	<input type="hidden" name="module" value="AsterLink">
	<input type="hidden" name="campaignConfig" value="true">
	<input type="hidden" name="action">
	<input type="hidden" name="return_module" value="{$RETURN_MODULE}">
	<input type="hidden" name="return_action" value="{$RETURN_ACTION}">
	<input type="hidden" name="source_form" value="config" />

<table width="100%" cellpadding="0" cellspacing="0" border="0">
	<tr>

		<td>
			<input title="{$APP.LBL_SAVE_BUTTON_TITLE}" accessKey="{$APP.LBL_SAVE_BUTTON_KEY}" class="button primary" onclick="this.form.action.value='Save';return verify_data(this);" type="submit" name="button" id="btn_save" value=" {$APP.LBL_SAVE_BUTTON_LABEL} ">
			<input title="{$APP.LBL_CANCEL_BUTTON_TITLE}" accessKey="{$APP.LBL_CANCEL_BUTTON_KEY}" class="button" onclick="this.form.action.value='{$RETURN_ACTION}'; this.form.module.value='{$RETURN_MODULE}';" type="submit" name="button" value=" {$APP.LBL_CANCEL_BUTTON_LABEL} ">
		</td>
		<td align="right" nowrap>
			<span class="required">{$APP.LBL_REQUIRED_SYMBOL}</span> {$APP.NTC_REQUIRED}
		</td>
	</tr>
</table>

<div id="EditView_tabs">
	<div class="panel-content">
		<div class="panel panel-default">
			<div class="panel-heading ">
				<a class="" role="button" data-toggle="collapse-edit" aria-expanded="false">
					<div class="col-xs-10 col-sm-11 col-md-11">
						{$MOD.LBL_CONFIG_ENDPOINT}
					</div>
				</a>
			</div>
			<div class="panel-body">
				<div class="tab-content">
					<table width="100%" border="0" cellspacing="0" cellpadding="0" class="edit view">
						<tr>
							<td width="15%" scope="row">
								{$MOD.LBL_CONFIG_TOKEN}
								{sugar_help text=$MOD.LBL_CONFIG_TOKEN_TT}
								<span class="required">{$APP.LBL_REQUIRED_SYMBOL}</span>
							</td>
							<td>
								<input name='endpoint_token' tabindex='1' type="text" value="{$ENDPOINT_TOKEN}" style="width: 40%;">
							</td>
						</tr>
						<tr>
							<td width="15%" scope="row">
								{$MOD.LBL_CONFIG_URL}
								{sugar_help text=$MOD.LBL_CONFIG_URL_TT}
							</td>
							<td>
								<input name="endpoint_url" tabindex="1" maxlength="128" type="text" placeholder="http://localhost:5678" value="{$ENDPOINT_URL}" style="width: 40%;">
							</td>
						</tr>
						<tr>
							<td width="15%" scope="row">
								{$MOD.LBL_CONFIG_PROXY}
								{sugar_help text=$MOD.LBL_CONFIG_PROXY_TT}
							</td>
							<td style="padding: 8px 0;">
								<input name="proxy_enabled" type="checkbox" value="true" {if $PROXY_ENABLED}checked{/if}/>
							</td>
						</tr>
					</table>
				</div>
			</div>
		</div>

		<div class="panel panel-default">
			<div class="panel-heading ">
				<a class="" role="button" data-toggle="collapse-edit" aria-expanded="false">
					<div class="col-xs-10 col-sm-11 col-md-11">
						{$MOD.LBL_CONFIG_RELS}
					</div>
				</a>
			</div>
			<div class="panel-body">
				<div class="tab-content">
					<table width="100%" border="0" cellspacing="0" cellpadding="0" class="table">
						<tr>
							<td scope="row" style="font-weight: bold; width: 15%;">
								{$MOD.LBL_CONFIG_RONCE}
								{sugar_help text=$MOD.LBL_CONFIG_RONCE_TT}
							</td>
							<td>
								<input name="relate_once" type="checkbox" value="true" {if $RELATE_ONCE}checked{/if}/>
							</td>
						</tr>
					</table>

					<table width="100%" border="0" cellspacing="0" cellpadding="0" class="table">
						<thead>
							<tr>
								<th width="20%">
									{$MOD.LBL_CONFIG_MODULE}
									{sugar_help text=$MOD.LBL_CONFIG_MODULE_TT}
								</th>
								<th width="30%">
									{$MOD.LBL_CONFIG_NFIELD}
									{sugar_help text=$MOD.LBL_CONFIG_NFIELD_TT}
									<span class="required">{$APP.LBL_REQUIRED_SYMBOL}</span>
								</th>
								<th width="30%">
									{$MOD.LBL_CONFIG_PFILED}
									{sugar_help text=$MOD.LBL_CONFIG_PFIELD_TT}
									<span class="required">{$APP.LBL_REQUIRED_SYMBOL}</span>
								</th>
								<th width="10%" class="text-center">
									{$MOD.LBL_CONFIG_SCREATE}
									{sugar_help text=$MOD.LBL_CONFIG_SCREATE_TT}
									<span class="required"></span>
								</th>
								<th width="10%"></th>
							</tr>
						</thead>
						<tbody id="relTbody">
							{foreach from=$REL_CONFIG item=rel key=i}
							<tr>
								<td>
									<select name="rel[{$i}]" data-id="{$i}" onchange="onRelChange(this)">
										{html_options options=$REL_MODULES|@array_column:title:id selected=$rel.rel_name}
									</select>
								</td>
								<td>
									<select name="rel_name_field[{$i}]">
										{html_options options=$REL_MODULES[$rel.rel_name].name_fields selected=$rel.name_field}
									</select>
								</td>
								<td>
									<select name="rel_phone_fields[{$i}][]" multiple>
										{html_options options=$REL_MODULES[$rel.rel_name].phone_fields selected=$rel.phone_fields}
									</select>
								</td>
								<td style="text-align: center;">
									<input type="checkbox" name="rel_show_create[{$i}]" {if $rel.show_create}checked{/if}/>
								</td>
								<td>
									<button class="button" onclick="delRel(this, event)">{$APP.LBL_DELETE_BUTTON}</button>
								</td>
							</tr>
							{/foreach}
						</tbody>
						<tfoot>
							<tr>
								<td></td>
								<td colspan="3">
									<button class="button" onclick="addNewRel(event)">{$APP.LBL_ADD_BUTTON}</button>
								</td>
							</tr>
						</tfoot>
					</table>
				</div>
			</div>
		</div>
	</div>
	<div class="panel-content">
		<div class="panel panel-default">
			<div class="panel-heading ">
				<a class="" role="button" data-toggle="collapse-edit" aria-expanded="false">
					<div class="col-xs-10 col-sm-11 col-md-11">
						{$MOD.LBL_CONFIG_UTILS}
					</div>
				</a>
			</div>
			<div class="panel-body">
				<div class="tab-content">
					<input type="submit" name="modify_duration" class="button" value="{$MOD.LBL_CONFIG_MDUR}" onclick="this.form.action.value='Save'" />	
				</div>
			</div>
		<div>
	</div>
</div>
</form>

<script type="text/javascript">
let relModuleFields = {literal}{{/literal}
{foreach from=$REL_MODULES key=module item=data}
	'{$module}': {literal}{{/literal}
		nameFields: `{html_options options=$data.name_fields}`,
		phoneFields: `{html_options options=$data.phone_fields}`,
	{literal}}{/literal},
{/foreach}
{literal}}{/literal};

let relLastId = {if $REL_CONFIG|is_array}{$REL_CONFIG|@count}{else}0{/if};

{literal}
function addNewRel(e) {
	e.preventDefault();

	let lastTr = $('#relTbody > tr:last-child');
	let id = relLastId;

	let tr = $('<tr />', {
		data: {id}
	});

	$('<td />', {
		html: `
		<select 
			data-id="${id}"
			name="rel[${id}]"
			onchange="onRelChange(this)"
			style="width: 100%"
		>
			{/literal}{html_options options=$REL_MODULES|@array_column:title:id}{literal}
		</select>
	`
	}).appendTo(tr);

	$('<td />', {
		html: `
		<select name="rel_name_field[${id}]" style="width: 100%">
		</select>
	`
	}).appendTo(tr);

	$('<td />', {
		html: `
		<select name="rel_phone_fields[${id}][]" style="width: 100%" multiple>
		</select>
	`
	}).appendTo(tr);

	$('<td />', {
		class: 'text-center',
		html: `
		<input type="checkbox" name="rel_show_create[${id}]" />
	`
	}).appendTo(tr);

	$('<td />', {
		html: `
		<button class="button" onclick="delRel(this, event)">Delete</button>
	`
	}).appendTo(tr);

	tr.appendTo($('#relTbody'));

	relLastId = id + 1;

	onRelChange($('[name="rel['+id+']"]'));
}

function onRelChange(sel) {
	let id = $(sel).data('id');

	let relModuleOption = relModuleFields[$(sel).val()];

	$('[name="rel_name_field['+id+']"]').html(relModuleOption.nameFields);
	$('[name="rel_phone_fields['+id+'][]"]').html(relModuleOption.phoneFields);
}

function delRel(btn) {
	$(btn).parent().parent().remove();

}
{/literal}
</script>